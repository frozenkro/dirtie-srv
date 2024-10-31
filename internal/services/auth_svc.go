package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type EmailSender interface {
	SendEmail(ctx context.Context, emailAddress string, subject string, body string) error
}

type HtmlParser interface {
	ReadFile(ctx context.Context, path string) (*template.Template, error)
	ReplaceVars(ctx context.Context, data any, tmp *template.Template) ([]byte, error)
	ReplaceAndWrite(ctx context.Context, data any, tmp *template.Template, w http.ResponseWriter) error
}

type UserReader interface {
	GetUser(ctx context.Context, userId int32) (sqlc.User, error)
	GetUserFromEmail(ctx context.Context, email string) (sqlc.User, error)
}
type UserWriter interface {
	CreateUser(ctx context.Context, email string, pwHash []byte, name string) (sqlc.User, error)
	ChangePassword(ctx context.Context, userId int32, pwHash []byte) error
	UpdateLastLoginTime(ctx context.Context, userId int32) error
}

type SessionReader interface {
	GetSession(ctx context.Context, token string) (sqlc.Session, error)
}
type SessionWriter interface {
	CreateSession(ctx context.Context, userId int32, token string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, token string) error
	DeleteUserSessions(ctx context.Context, userId int32) error
}

type PwResetReader interface {
	GetPwResetToken(ctx context.Context, token string) (sqlc.PwResetToken, error)
}
type PwResetWriter interface {
	CreatePwResetToken(ctx context.Context, userId int32, token string, expiresAt time.Time) error
	DeletePwResetToken(ctx context.Context, token string) error
	DeleteUserPwResetTokens(ctx context.Context, userId int32) error
}

type AuthSvc struct {
	userReader    UserReader
	userWriter    UserWriter
	sessionReader SessionReader
	sessionWriter SessionWriter
	pwResetReader PwResetReader
	pwResetWriter PwResetWriter
	htmlParser    HtmlParser
	emailSender   EmailSender
}

var (
	ErrInvalidToken    = fmt.Errorf("Invalid auth token")
	ErrExpiredToken    = fmt.Errorf("Auth token expired")
	ErrUserExists      = fmt.Errorf("User Email already exists")
	ErrNoUser          = fmt.Errorf("User not found")
	ErrInvalidPassword = fmt.Errorf("Invalid Password")
)

func NewAuthSvc(userReader UserReader,
	userWriter UserWriter,
	sessionReader SessionReader,
	sessionWriter SessionWriter,
	pwResetReader PwResetReader,
	pwResetWriter PwResetWriter,
	htmlParser HtmlParser,
	emailSender EmailSender) *AuthSvc {

	return &AuthSvc{
		userReader:    userReader,
		userWriter:    userWriter,
		sessionReader: sessionReader,
		sessionWriter: sessionWriter,
		pwResetReader: pwResetReader,
		pwResetWriter: pwResetWriter,
		htmlParser:    htmlParser,
		emailSender:   emailSender,
	}
}

type ReplaceVars struct {
	Username  string
	ResetLink string
}

func (s *AuthSvc) CreateUser(ctx context.Context, email string, password string, name string) (*sqlc.User, error) {

	existingUser, err := s.userReader.GetUserFromEmail(ctx, email)
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

	newUser, err := s.userWriter.CreateUser(ctx, email, pwHash, name)
	if err != nil {
		return nil, fmt.Errorf("Error CreateUser -> CreateUser: \n%w\n", err)
	}
	return &newUser, err
}

func (s *AuthSvc) Login(ctx context.Context, email string, password string) (string, error) {
	user, err := s.userReader.GetUserFromEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("Error Login -> GetUserFromEmail: \n%w\n", err)
	}

	err = bcrypt.CompareHashAndPassword(user.PwHash, []byte(password))
	if err != nil {
		return "", fmt.Errorf("Error Login -> CompareHashAndPassword: \n%w\n", ErrInvalidPassword)
	}

	err = s.userWriter.UpdateLastLoginTime(ctx, user.UserID)
	if err != nil {
		return "", fmt.Errorf("Error Login -> UpdateLastLoginTime: \n%w\n", err)
	}

	err = s.sessionWriter.DeleteUserSessions(ctx, user.UserID)
	if err != nil {
		return "", fmt.Errorf("Error Login -> DeleteUserSessions: \n%w\n", err)
	}

	token, expiresAt, err := createToken()
	if err != nil {
		return "", fmt.Errorf("Error Login -> createToken: \n%w\n", err)
	}

	err = s.sessionWriter.CreateSession(ctx, user.UserID, token, expiresAt)
	if err != nil {
		return "", fmt.Errorf("Error Login -> CreateSession: \n%w\n", err)
	}
	return token, nil
}

func (s *AuthSvc) ValidateToken(ctx context.Context, token string) (*sqlc.User, error) {
	session, err := s.sessionReader.GetSession(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("Error ValidateToken -> GetSession: \n%w\n", err)
	}

	if session.UserID < 1 {
		return nil, fmt.Errorf("ValidateToken - Validating token %v: \n%w\n", token, ErrInvalidToken)
	}

	if time.Now().After(session.ExpiresAt.Time) {
		return nil, fmt.Errorf("ValidateToken - Validating token %v: \n%w\n", token, ErrExpiredToken)
	}

	user, err := s.userReader.GetUser(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("Error ValidateToken -> GetUser: \n%w\n", err)
	}

	return &user, nil
}

func (s *AuthSvc) Logout(ctx context.Context, token string) error {
	session, err := s.sessionReader.GetSession(ctx, token)
	if err != nil {
		return fmt.Errorf("Error Logout -> GetSession: \n%w\n", err)
	}

	err = s.sessionWriter.DeleteUserSessions(ctx, session.UserID)
	if err != nil {
		return fmt.Errorf("Error Logout -> DeleteUserSessions: \n%w\n", err)
	}
	return nil
}

func (s *AuthSvc) ForgotPw(ctx context.Context, email string) error {
	// Find user
	user, err := s.userReader.GetUserFromEmail(ctx, email)
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

	err = s.pwResetWriter.DeleteUserPwResetTokens(ctx, userId)
	if err != nil {
		return fmt.Errorf("Error ForgotPw -> DeleteUserPwResetTokens: \n%w\n", err)
	}
	err = s.pwResetWriter.CreatePwResetToken(ctx, userId, token, expiresAt)
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
	res, err := s.pwResetReader.GetPwResetToken(ctx, token)
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

	s.userWriter.ChangePassword(ctx, userId, pwHash)
	err = s.pwResetWriter.DeleteUserPwResetTokens(ctx, userId)
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
