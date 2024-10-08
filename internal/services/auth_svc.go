package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/core/utils"
	"github.com/frozenkro/dirtie-srv/internal/db/repos"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthSvc struct {
	userRepo    repos.UserRepo
	sessionRepo repos.SessionRepo
	pwResetRepo repos.PwResetRepo
	htmlParser  utils.HtmlParser
	emailSender utils.EmailSender
}

var (
	ErrInvalidToken = fmt.Errorf("Invalid auth token")
	ErrExpiredToken = fmt.Errorf("Auth token expired")
	ErrUserExists   = fmt.Errorf("User Email already exists")
	ErrNoUser       = fmt.Errorf("User not found")
)

func NewAuthSvc(userRepo repos.UserRepo,
	sessionRepo repos.SessionRepo,
	pwResetRepo repos.PwResetRepo,
	htmlParser utils.HtmlParser,
	emailSender utils.EmailSender) *AuthSvc {

	return &AuthSvc{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		pwResetRepo: pwResetRepo,
		htmlParser:  htmlParser,
		emailSender: emailSender,
	}
}

type ReplaceVars struct {
	Username  string
	ResetLink string
}

func (s *AuthSvc) CreateUser(ctx context.Context, email string, password string, name string) (*sqlc.User, error) {

	existingUser, err := s.userRepo.GetUserFromEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("Error CreateUser -> GetUserFromEmail: \n%w\n", err)
	}
	if existingUser.UserID > 0 {
		return nil, fmt.Errorf("Error CreateUser - creating user with email %v: \n%w\n", email, ErrUserExists)
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return nil, fmt.Errorf("Error CreateUser -> GenerateFromPassword: \n%w\n", err)
	}

	newUser, err := s.userRepo.CreateUser(ctx, email, pwHash, name)
	if err != nil {
		return nil, fmt.Errorf("Error CreateUser -> CreateUser: \n%w\n", err)
	}
	return &newUser, err
}

func (s *AuthSvc) Login(ctx context.Context, email string, password string) (string, error) {
	user, err := s.userRepo.GetUserFromEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("Error Login -> GetUserFromEmail: \n%w\n", err)
	}

	err = bcrypt.CompareHashAndPassword(user.PwHash, []byte(password))
	if err != nil {
		return "", fmt.Errorf("Error Login -> CompareHashAndPassword: \n%w\n", err)
	}

	err = s.userRepo.UpdateLastLoginTime(ctx, user.UserID)
	if err != nil {
		return "", fmt.Errorf("Error Login -> UpdateLastLoginTime: \n%w\n", err)
	}

	err = s.sessionRepo.DeleteUserSessions(ctx, user.UserID)
	if err != nil {
		return "", fmt.Errorf("Error Login -> DeleteUserSessions: \n%w\n", err)
	}

	token, expiresAt, err := createToken()
	if err != nil {
		return "", fmt.Errorf("Error Login -> createToken: \n%w\n", err)
	}

	err = s.sessionRepo.CreateSession(ctx, user.UserID, token, expiresAt)
	if err != nil {
		return "", fmt.Errorf("Error Login -> CreateSession: \n%w\n", err)
	}
	return token, nil
}

func (s *AuthSvc) ValidateToken(ctx context.Context, token string) (*sqlc.User, error) {
	session, err := s.sessionRepo.GetSession(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("Error ValidateToken -> GetSession: \n%w\n", err)
	}

	if session.UserID < 1 {
		return nil, fmt.Errorf("ValidateToken - Validating token %v: \n%w\n", token, ErrInvalidToken)
	}

	if time.Now().After(session.ExpiresAt.Time) {
		return nil, fmt.Errorf("ValidateToken - Validating token %v: \n%w\n", token, ErrExpiredToken)
	}

	user, err := s.userRepo.GetUser(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("Error ValidateToken -> GetUser: \n%w\n", err)
	}

	return &user, nil
}

func (s *AuthSvc) Logout(ctx context.Context, token string) error {
	session, err := s.sessionRepo.GetSession(ctx, token)
	if err != nil {
		return fmt.Errorf("Error Logout -> GetSession: \n%w\n", err)
	}

	err = s.sessionRepo.DeleteUserSessions(ctx, session.UserID)
	if err != nil {
		return fmt.Errorf("Error Logout -> DeleteUserSessions: \n%w\n", err)
	}
	return nil
}

func (s *AuthSvc) ForgotPw(ctx context.Context, email string) error {
	// Find user
	user, err := s.userRepo.GetUserFromEmail(ctx, email)
	if err != nil {
		return err
	}
	if user.UserID <= 0 {
    return fmt.Errorf("No user found for email '%v': %w\n", email, ErrNoUser)
	}
	userId := user.UserID

	// create token
	token, expiresAt, err := createToken()
	if err != nil {
		return fmt.Errorf("Error ForgotPw -> createToken: \n%w\n", err)
	}

	err = s.pwResetRepo.DeleteUserPwResetTokens(ctx, userId)
	if err != nil {
		return fmt.Errorf("Error ForgotPw -> DeleteUserPwResetTokens: \n%w\n", err)
	}
	err = s.pwResetRepo.CreatePwResetToken(ctx, userId, token, expiresAt)
	if err != nil {
		return fmt.Errorf("Error ForgotPw -> CreatePwResetToken: \n%w\n", err)
	}

	// load and fill html template
	template, err := s.htmlParser.ReadFile(ctx, fmt.Sprintf("%vresetPwEmail.html", core.ASSETS_DIR))
	if err != nil {
		return fmt.Errorf("Error ForgotPw -> ReadFile: \n%w\n", err)
	}

	encToken := base64.URLEncoding.EncodeToString([]byte(token))
	resetLink := fmt.Sprintf("localhost:8080/pw/change?token=%v", encToken)
	vars := &ReplaceVars{Username: user.Name, ResetLink: resetLink}
	body, err := s.htmlParser.ReplaceVars(ctx, vars, template)
	if err != nil {
		return fmt.Errorf("Error ForgotPw -> ReplaceVars: \n%w\n", err)
	}

	// send email
	err = s.emailSender.SendEmail(ctx, user.Email, "Dirtie Password Reset Request", string(body))
	if err != nil {
		return fmt.Errorf("Error ForgotPw -> SendEmail: \n%w\n", err)
	}

	return nil
}

func (s *AuthSvc) ValidateForgotPwToken(ctx context.Context, encToken string) (int32, error) {
	// decode token
	bytes, err := base64.URLEncoding.DecodeString(encToken)
	if err != nil {
		return 0, fmt.Errorf("Error ValidateForgotPwToken -> DecodeString: \n%w\n", err)
	}
	token := string(bytes)

	// get token from db
	res, err := s.pwResetRepo.GetPwResetToken(ctx, token)
	if err != nil {
		return 0, fmt.Errorf("Error ValidateForgotPwToken -> GetPwResetToken: \n%w\n", err)
	}
	if res.UserID < 1 {
		return 0, fmt.Errorf("Error ValidateForgotPwToken - '%v': \n%w\n", token, ErrInvalidToken)
	}

	if time.Now().After(res.ExpiresAt.Time) {
		return 0, fmt.Errorf("Error ValidateForgotPwToken - '%v': \n%w\n", token, ErrExpiredToken)
	}

	// return user id
	return res.UserID, nil
}

func (s *AuthSvc) ChangePw(ctx context.Context, encToken string, newPw string) error {
	userId, err := s.ValidateForgotPwToken(ctx, encToken)
	if err != nil {
		return fmt.Errorf("Error ChangePw -> ValidateForgotPwToken: \n%w\n", err)
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(newPw), 10)
	if err != nil {
		return fmt.Errorf("Error ChangePw -> GenerateFromPassword: \n%w\n", err)
	}

	s.userRepo.ChangePassword(ctx, userId, pwHash)
	err = s.pwResetRepo.DeleteUserPwResetTokens(ctx, userId)
	if err != nil {
		return fmt.Errorf("Error ChangePw - An error occurred after successful password change: \n%w\n", err)
	}

	return nil
}

func createToken() (string, time.Time, error) {

	token := uuid.NewString()
	dur, err := time.ParseDuration("1h")
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(dur)

	return token, expiresAt, nil
}
