package int_tst

import (
	"context"
	"fmt"
	"os"
  "time"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var ( 
  setupComplete bool = false
  TestUser sqlc.User
  TestSession sqlc.Session
  TestPwResetToken sqlc.PwResetToken
  TestDevice sqlc.Device
  TestProvStg sqlc.ProvisionStaging
)

func SetupTests() {
  if setupComplete {
    return
  }

  core.SetupTestEnv()
  setupDb()

  setupComplete = true
}

func ConnectDb() *pgx.Conn {
	connstr := fmt.Sprintf("postgres://%v:%v@%v/%v",
		core.POSTGRES_USER,
		core.POSTGRES_PASSWORD,
		core.POSTGRES_SERVER,
		core.POSTGRES_DB)
  db, err := pgx.Connect(context.Background(), connstr)
  if err != nil {
    panic(fmt.Errorf("Error connecting to test db: %w", err))
  }

  return db
}
func setupDb() {
  db := ConnectDb()
  defer db.Close(context.Background())

  schema, err := os.ReadFile(core.ProjectRootDir() + "/internal/db/sqlc/schema.sql")
  if err != nil {
    panic(fmt.Errorf("Error reading schema.sql: %w", err))
  }

  if _, err := db.Exec(context.Background(), string(schema)); err != nil {
    panic(fmt.Errorf("Error executing schema.sql: %w", err))
  }

  setupData(db)
}

func setupData(db *pgx.Conn) {
  TestUser.Name = "Test User"
  TestUser.Email = "test@email.test"
  TestUser.PwHash = []byte("pwhash")
  TestUser.UserID = 1
  userSql := "INSERT INTO users (email, name, pw_hash) VALUES ($1, $2, $3)"
  if _, err := db.Exec(context.Background(), userSql, TestUser.Email, TestUser.Name, TestUser.PwHash); err != nil {
    panic(fmt.Errorf("Error creating test user record: %w", err))
  }
  
  TestSession.Token = "testtoken"
  TestSession.UserID = 1
  duration, _ := time.ParseDuration("2h")
  TestSession.ExpiresAt.Time = time.Now().Add(duration)
  TestSession.ExpiresAt.Valid = true
  TestSession.SessionID = 1
  sessionSql := "INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, $3)"
  if _, err := db.Exec(context.Background(), sessionSql, TestSession.UserID, TestSession.Token, TestSession.ExpiresAt); err != nil {
    panic(fmt.Errorf("Error creating test session record: %w", err))
  }
  
  TestPwResetToken.Token = "testpwresettoken"
  TestPwResetToken.UserID = 1
  TestPwResetToken.ExpiresAt.Time = time.Now().Add(duration)
  TestPwResetToken.ExpiresAt.Valid = true
  TestPwResetToken.PwResetID = 1
  pwRstSql := "INSERT INTO pw_reset_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)"
  if _, err := db.Exec(context.Background(), pwRstSql, TestPwResetToken.UserID, TestPwResetToken.Token, TestPwResetToken.ExpiresAt); err != nil {
    panic(fmt.Errorf("Error creating test pw reset token record: %w", err))
  }

  TestDevice.MacAddr = pgtype.Text{
    String: "testmacaddr",
    Valid: true,
  }
  TestDevice.DeviceID = 1
  TestDevice.UserID = 1
  TestDevice.DisplayName = pgtype.Text{
    String: "testDeviceDisplay",
    Valid: true,
  }
  deviceSql := "INSERT INTO devices (user_id, mac_addr, display_name) VALUES ($1, $2, $3)"
  if _, err := db.Exec(context.Background(), deviceSql, TestDevice.UserID, TestDevice.MacAddr, TestDevice.DisplayName); err != nil {
    panic(fmt.Errorf("Error creating test device record: %w", err))
  }

  TestProvStg.Contract.String = "testprvstgcontract"
  TestProvStg.Contract.Valid = true
  TestProvStg.DeviceID = 1
  prvSql := "INSERT INTO provision_staging (device_id, contract) VALUES ($1, $2)"
  if _, err := db.Exec(context.Background(), prvSql, TestProvStg.DeviceID, TestProvStg.Contract); err != nil {
    panic(fmt.Errorf("Error creating test provision staging record: %w", err))
  }
}
