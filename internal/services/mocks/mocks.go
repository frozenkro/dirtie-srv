package mocks

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/db"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockUserReader struct {
	*mock.Mock
}

type MockUserWriter struct {
	*mock.Mock
}

type MockSessionReader struct {
	*mock.Mock
}
type MockSessionWriter struct {
	*mock.Mock
}

type MockPwResetReader struct {
	*mock.Mock
}
type MockPwResetWriter struct {
	*mock.Mock
}

type MockHtmlParser struct {
	*mock.Mock
}

type MockEmailSender struct {
	*mock.Mock
}

type MockDeviceDataRetriever struct {
	*mock.Mock
}

type MockDeviceDataRecorder struct {
	*mock.Mock
}

type MockDeviceGetter struct {
	*mock.Mock
}

type MockDeviceReader struct {
	*mock.Mock
}
type MockDeviceWriter struct {
	*mock.Mock
}

type MockPrvStgReader struct {
	*mock.Mock
}
type MockPrvStgWriter struct {
	*mock.Mock
}
type MockUserCtxReader struct {
	*mock.Mock
}

// Implement UserRepo interface methods for MockUserRepo
func (m MockUserReader) GetUserFromEmail(ctx context.Context, email string) (sqlc.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m MockUserWriter) CreateUser(ctx context.Context, email string, pwHash []byte, name string) (sqlc.User, error) {
	args := m.Called(ctx, email, pwHash, name)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m MockUserWriter) UpdateLastLoginTime(ctx context.Context, userID int32) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m MockUserReader) GetUser(ctx context.Context, userID int32) (sqlc.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m MockUserWriter) ChangePassword(ctx context.Context, userId int32, pwHash []byte) error {
	args := m.Called(ctx, userId, pwHash)
	return args.Error(0)
}

// Implement SessionRepo interface methods for MockSessionRepo
func (m MockSessionWriter) CreateSession(ctx context.Context, userID int32, token string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Error(0)
}

func (m MockSessionReader) GetSession(ctx context.Context, token string) (sqlc.Session, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(sqlc.Session), args.Error(1)
}

func (m MockSessionWriter) DeleteSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m MockSessionWriter) DeleteUserSessions(ctx context.Context, userId int32) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}

func (m MockPwResetWriter) CreatePwResetToken(ctx context.Context, userId int32, token string, expiresAt time.Time) error {
	args := m.Called(ctx, userId, token, expiresAt)
	return args.Error(0)
}
func (m MockPwResetReader) GetPwResetToken(ctx context.Context, token string) (sqlc.PwResetToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(sqlc.PwResetToken), args.Error(1)
}
func (m MockPwResetWriter) DeletePwResetToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}
func (m MockPwResetWriter) DeleteUserPwResetTokens(ctx context.Context, userId int32) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}

func (m MockHtmlParser) ReadFile(ctx context.Context, path string) (*template.Template, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(*template.Template), args.Error(1)
}

func (m MockHtmlParser) ReplaceVars(ctx context.Context, vars any, tmp *template.Template) ([]byte, error) {
	args := m.Called(ctx, vars, tmp)
	return args.Get(0).([]byte), args.Error(1)
}

func (m MockHtmlParser) ReplaceAndWrite(ctx context.Context, data any, tmp *template.Template, w http.ResponseWriter) error {
	args := m.Called(ctx, data, tmp, w)
	return args.Error(0)
}

func (m MockEmailSender) SendEmail(ctx context.Context, emailAddress string, subject string, body string) error {
	args := m.Called(ctx, emailAddress, subject, body)
	return args.Error(0)
}

func (m MockDeviceDataRetriever) GetLatestValue(ctx context.Context, deviceId int, measurementKey string) (db.DeviceDataPoint, error) {
	args := m.Called(ctx, deviceId, measurementKey)
	return args.Get(0).(db.DeviceDataPoint), args.Error(1)
}

func (m MockDeviceDataRetriever) GetValuesRange(
	ctx context.Context,
	deviceId int,
	measurementKey string,
	start time.Time,
	end time.Time) ([]db.DeviceDataPoint, error) {
	args := m.Called(ctx, deviceId, measurementKey, start, end)
	return args.Get(0).([]db.DeviceDataPoint), args.Error(1)
}

func (m MockDeviceDataRecorder) Record(
	ctx context.Context,
	deviceId int,
	measurementKey string,
	value int64) error {
	args := m.Called(ctx, deviceId, measurementKey, value)
	return args.Error(0)
}

func (m MockDeviceGetter) GetDeviceByMacAddress(ctx context.Context, macAddr string) (sqlc.Device, error) {
	args := m.Called(ctx, macAddr)
	return args.Get(0).(sqlc.Device), args.Error(1)
}

func (m MockDeviceReader) GetDeviceByMacAddress(ctx context.Context, macAddr string) (sqlc.Device, error) {
	args := m.Called(ctx, macAddr)
	return args.Get(0).(sqlc.Device), args.Error(1)
}

func (m MockDeviceReader) GetDevicesByUser(ctx context.Context, userId int32) ([]sqlc.Device, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]sqlc.Device), args.Error(1)
}

func (m MockDeviceWriter) CreateDevice(ctx context.Context, userId int32, displayName string) (sqlc.Device, error) {
	args := m.Called(ctx, userId, displayName)
	return args.Get(0).(sqlc.Device), args.Error(1)
}

func (m MockDeviceWriter) RenameDevice(ctx context.Context, deviceId int32, displayName string) error {
	args := m.Called(ctx, deviceId, displayName)
	return args.Error(0)
}

func (m MockDeviceWriter) UpdateDeviceMacAddress(ctx context.Context, deviceId int32, macAddr string) error {
	args := m.Called(ctx, deviceId, macAddr)
	return args.Error(0)
}

func (m MockPrvStgWriter) CreateProvisionStaging(ctx context.Context, deviceId int32, contract string) error {
	args := m.Called(ctx, deviceId, contract)
	return args.Error(0)
}

func (m MockPrvStgWriter) DeleteProvisionStaging(ctx context.Context, deviceId int32) error {
	args := m.Called(ctx, deviceId)
	return args.Error(0)
}

func (m MockPrvStgReader) GetProvisionStagingByContract(ctx context.Context, contract string) (sqlc.ProvisionStaging, error) {
	args := m.Called(ctx, contract)
	return args.Get(0).(sqlc.ProvisionStaging), args.Error(1)
}

func (m MockUserCtxReader) GetUser(ctx context.Context) (sqlc.User, error) {
	args := m.Called(ctx)
	return args.Get(0).(sqlc.User), args.Error(1)
}
