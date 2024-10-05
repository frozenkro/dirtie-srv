package api

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
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

func setupDb() {
	connstr := fmt.Sprintf("postgres://%v:%v@%v/%v",
		core.POSTGRES_USER,
		core.POSTGRES_PASSWORD,
		core.POSTGRES_SERVER,
		core.POSTGRES_DB)
  db, err := sql.Open("postgres", connstr)
  if err != nil {
    panic(fmt.Errorf("Error connecting to test db: %w", err))
  }
  defer db.Close()

  schema, err := os.ReadFile("../db/sqlc/schema.sql")
  if err != nil {
    panic(fmt.Errorf("Error reading schema.sql: %w", err))
  }

  if _, err := db.Exec(string(schema)); err != nil {
    panic(fmt.Errorf("Error executing schema.sql: %w", err))
  }

  setupData(db)
}

func setupData(db *sql.DB) {
  TestUser.Name = "Test User"
  TestUser.Email = "test@email.test"
  TestUser.PwHash = []byte("pwhash")
  TestUser.UserID = 1
  userSql := fmt.Sprintf("INSERT INTO users (email, name, pw_hash) VALUES (%v, %v, %v)", TestUser.Name, TestUser.Email, TestUser.PwHash)
  if _, err := db.Exec(userSql); err != nil {
    panic(fmt.Errorf("Error creating test user record: %w", err))
  }
  
  TestSession.Token = "testtoken"
  TestSession.UserID = 1
  TestSession.SessionID = 1
  sessionSql := fmt.Sprintf("INSERT INTO sessions (user_id, token) VALUES (%v, %v)", TestSession.UserID, TestSession.Token)
  if _, err := db.Exec(sessionSql); err != nil {
    panic(fmt.Errorf("Error creating test session record: %w", err))
  }
  
  TestPwResetToken.Token = "testpwresettoken"
  TestPwResetToken.UserID = 1
  TestPwResetToken.PwResetID = 1
  pwRstSql := fmt.Sprintf("INSERT INTO pw_reset_tokens (user_id, token) VALUES (%v, %v)", TestPwResetToken.UserID, TestPwResetToken.Token)
  if _, err := db.Exec(pwRstSql); err != nil {
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
  deviceSql := fmt.Sprintf("INSERT INTO devices (user_id, mac_addr, display_name) VALUES (%v, %v, %v)", TestDevice.UserID, TestDevice.MacAddr, TestDevice.DisplayName)
  if _, err := db.Exec(deviceSql); err != nil {
    panic(fmt.Errorf("Error creating test device record: %w", err))
  }

  TestProvStg.Contract.String = "testprvstgcontract"
  TestProvStg.Contract.Valid = true
  TestProvStg.DeviceID = 1
  prvSql := fmt.Sprintf("INSERT INTO provision_staging (device_id, contract) VALUES (%v, %v)", TestProvStg.DeviceID, TestProvStg.Contract)
  if _, err := db.Exec(prvSql); err != nil {
    panic(fmt.Errorf("Error creating test provision staging record: %w", err))
  }
}
