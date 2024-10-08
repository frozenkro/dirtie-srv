package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/core/int_tst"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/di"
)

func TestCreateUser_CreatesUser(t *testing.T) {
  core.SetupTestEnv()
  int_tst.SetupTests()
  deps := di.NewDeps()
  server := httptest.NewServer(createUserHandler(deps.AuthSvc))
  defer server.Close()

  userArgs := CreateUserArgs{
    Email: "createusertest@email.com",
    Password: "createuserpassword",
    Name: "Test User",
  }
  userBytes, err := json.Marshal(userArgs)
  if err != nil {
    t.Fatalf("Error encoding request body: %v", err)
  }

  resp, err := http.Post(server.URL+"/users", "application/json", bytes.NewBuffer(userBytes))
  if err != nil {
    t.Errorf("API client returned error: %v", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    t.Errorf("API status code unexpected: %v", resp.StatusCode)
  }
  
  db := int_tst.ConnectDb()
  row, err := db.Query(
    context.Background(),
    "SELECT * FROM users WHERE email = $1", 
    userArgs.Email,
  )

  if !row.Next() {
    t.Errorf("No rows found")
  }

  user := sqlc.User{}
  err = row.Scan(&user.UserID, &user.Email, &user.Name, &user.PwHash, &user.CreatedAt, &user.LastLogin)
  if err != nil {
    t.Fatalf("Error converting inserted row to user obj: %v", err)
  }

  if row.Next() {
    t.Errorf("Multiple rows found")
  }

  if user.Name != userArgs.Name {
    t.Errorf("Inserted user Name mismatch")
  }
}

func TestCreateUser_WhenUserEmailExists_ReturnsUserExistsError(t *testing.T) {
  core.SetupTestEnv()
  int_tst.SetupTests()
  deps := di.NewDeps()
  server := httptest.NewServer(createUserHandler(deps.AuthSvc))
  defer server.Close()

  userArgs := CreateUserArgs{
    Email: int_tst.TestUser.Email,
    Password: "createuserpassword",
    Name: "Test User",
  }
  userBytes, err := json.Marshal(userArgs)
  if err != nil {
    t.Fatalf("Error encoding request body: %v", err)
  }

  resp, err := http.Post(server.URL+"/users", "application/json", bytes.NewBuffer(userBytes))
  if err != nil {
    t.Errorf("API client returned error: %v", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusInternalServerError {
    t.Errorf("API status code unexpected: %v", resp.StatusCode)
  }
}
