package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/core/int_testing"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/di"
)

func TestCreateUser_CreatesUser(t *testing.T) {
  core.SetupTestEnv()
  int_testing.SetupTests()
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
    t.Errorf("API returned error: %v", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    t.Errorf("API status code unexpected: %v", resp.StatusCode)
  }
  
  db := int_testing.ConnectDb()
  row, err := db.Query(
    fmt.Sprintf("SELECT * FROM users WHERE email = %v", userArgs.Email),
  )

  if row.Next() {
    t.Errorf("Multiple records found")
  }

  user := sqlc.User{}
  err = row.Scan(user.Email, user.PwHash, user.Name, user.CreatedAt, user.LastLogin)
  if err != nil {
    t.Fatalf("Error converting inserted row to user obj: %v", err)
  }

  if user.Name != userArgs.Name {
    t.Errorf("Inserted user Name mismatch")
  }
}
