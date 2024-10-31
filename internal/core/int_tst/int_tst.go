package int_tst

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/core"
	drt_db "github.com/frozenkro/dirtie-srv/internal/db"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	initSetup        bool = false
	TestUser          sqlc.User
	TestSession       sqlc.Session
	TestPwResetToken  sqlc.PwResetToken
	TestDevice        sqlc.Device
  TestProvStgDevice sqlc.Device
	TestProvStg       sqlc.ProvisionStaging
)

func TestContext(t *testing.T) context.Context {
	return context.WithValue(
		context.Background(),
		"testdb",
		strings.ToLower(t.Name()),
	)
}

func SetupTests(ctx context.Context, t *testing.T) *pgx.Conn {
	if !initSetup {
		core.SetupTestEnv()
		initSetup = true
	}

	db := connectDb(ctx, t)
	setupDb(db)
	return db
}

func connectDb(ctx context.Context, t *testing.T) *pgx.Conn {
	dbName := ctx.Value("testdb")
	if dbName == nil || dbName == "" {
		t.Fatalf("Test '%v' does not pass context with testdb value to connectDb. Use TestContext(t)", t.Name())
	}

	mconnstr := fmt.Sprintf("postgres://%v:%v@%v/%v",
		core.POSTGRES_USER,
		core.POSTGRES_PASSWORD,
		core.POSTGRES_SERVER,
		"postgres")
	maintDb, err := pgx.Connect(ctx, mconnstr)
	if err != nil {
		t.Fatalf("Error connecting to test maintenance db: %v", err)
	}
	defer maintDb.Close(ctx)

	exists := 0
	err = maintDb.QueryRow(ctx, "SELECT 1 FROM pg_database WHERE datname = $1", dbName).Scan(&exists)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("Error occurred when checking for test db '%v' existence: %v", dbName, err)
	}
	if exists == 1 {
		t.Fatalf("Cannot create db for test '%v'. Test name is not unique", dbName)
	}

	_, err = maintDb.Exec(ctx, fmt.Sprintf("CREATE DATABASE %v", dbName))
	if err != nil {
		t.Fatalf("Error occurred when creating test db '%v': %v", dbName, err)
	}

	connstr := fmt.Sprintf("postgres://%v:%v@%v/%v",
		core.POSTGRES_USER,
		core.POSTGRES_PASSWORD,
		core.POSTGRES_SERVER,
		dbName)
	db, err := pgx.Connect(ctx, connstr)
	if err != nil {
		t.Fatalf("Error connecting to freshly created test db '%v': %v", dbName, err)
	}

	return db
}

func setupDb(db *pgx.Conn) {
	schema := drt_db.SchemaSql

	if _, err := db.Exec(context.Background(), string(schema)); err != nil {
		panic(fmt.Errorf("Error executing schema.sql: %w", err))
	}

	setupData(db)
}

func setupData(db *pgx.Conn) {
	TestUser.Name = "Test User"
	TestUser.Email = "test@email.test"
	// "testpw"
	TestUser.PwHash = []byte("$2a$10$sYZTH/eivjOREKa/ehkWQ.7SjbsLDJEfoTzpsjKIwVafkijloIndi")
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
		Valid:  true,
	}
	TestDevice.DeviceID = 1
	TestDevice.UserID = 1
	TestDevice.DisplayName = pgtype.Text{
		String: "testDeviceDisplay",
		Valid:  true,
	}
	deviceSql := "INSERT INTO devices (user_id, mac_addr, display_name) VALUES ($1, $2, $3)"
	if _, err := db.Exec(context.Background(), deviceSql, TestDevice.UserID, TestDevice.MacAddr, TestDevice.DisplayName); err != nil {
		panic(fmt.Errorf("Error creating test device record: %w", err))
	}

  TestProvStgDevice.DisplayName = pgtype.Text{
    String: "testprvstgdevice",
    Valid:  true,
  }
  TestProvStgDevice.UserID = 1
  prvstgdvcSql := `INSERT INTO devices (user_id, display_name) 
    VALUES ($1, $2)
    RETURNING device_id;`
  err := db.QueryRow(context.Background(), prvstgdvcSql, TestProvStgDevice.UserID, TestProvStgDevice.DisplayName).Scan(&TestProvStg.DeviceID) 
  if err != nil {
    panic(fmt.Errorf("Error creating prv stg test device record: %w", err))
  }

	TestProvStg.Contract.String = "testprvstgcontract"
	TestProvStg.Contract.Valid = true
	prvSql := "INSERT INTO provision_staging (device_id, contract) VALUES ($1, $2)"
	if _, err := db.Exec(context.Background(), prvSql, TestProvStg.DeviceID, TestProvStg.Contract); err != nil {
		panic(fmt.Errorf("Error creating test provision staging record: %w", err))
	}
}
