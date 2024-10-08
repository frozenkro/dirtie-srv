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
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
  core.SetupTestEnv()
  int_tst.SetupTests()
  db := int_tst.ConnectDb()
  deps := di.NewDeps()
  server := httptest.NewServer(createUserHandler(deps.AuthSvc))
  defer server.Close()

  t.Run("Success", func(t *testing.T) {

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

    assert.False(t, row.Next())
    assert.Equal(t, user.Name, userArgs.Name)
  })

  t.Run("UserAlreadyExists", func(t *testing.T) {
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

    assert.Equal(t, resp.StatusCode, http.StatusInternalServerError)
  })
}
