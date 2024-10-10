package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/core/int_tst"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/di"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
  ctx := context.Background()
  db := int_tst.SetupTests()
  defer db.Close(ctx)

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

func TestLogin(t *testing.T) {
  ctx := context.Background()
  db := int_tst.SetupTests()
  defer db.Close(ctx)

  deps := di.NewDeps()
  server := httptest.NewServer(loginHandler(deps.AuthSvc))
  defer server.Close()

  t.Run("Success", func(t *testing.T) {
    loginArgs := LoginArgs{
      Email: int_tst.TestUser.Email,
      Password: "testpw",
    }

    loginBytes, err := json.Marshal(loginArgs)
    if err != nil {
      t.Errorf("Error encoding request body: %v", err)
    }

    resp, err := http.Post(server.URL+"/login", "application/json", bytes.NewBuffer(loginBytes))
    if err != nil {
      t.Errorf("API client returned error: %v", err)
    }
    
    assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "Invalid credentials error")
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    cookies := resp.Cookies()
    authIdx := slices.IndexFunc(cookies, func(c *http.Cookie) bool { return c.Name == core.AUTH_COOKIE_NAME })
    assert.NotEqual(t, -1, authIdx) 

    if authIdx >= 0 {
      assert.NotNil(t, cookies[authIdx].Value)
      assert.Nil(t, cookies[authIdx].Valid())
    }

    // updates LastLogin
    user := sqlc.User{Email: loginArgs.Email}
    userR, err := db.Query(ctx, `SELECT user_id, last_login
      FROM users
      WHERE email = $1`, 
      loginArgs.Email)
    if err != nil {
      t.Fatalf("Error querying db for test results: %v", err)
    }
    assert.True(t, userR.Next())
    userR.Scan(&user.UserID, &user.LastLogin)
    recentTime := time.Now().Add(-2 * time.Minute)
    assert.True(t, recentTime.Before(user.LastLogin.Time))
    assert.False(t, userR.Next())

    // creates session
    session := sqlc.Session{ UserID: user.UserID }
    sessionR, err := db.Query(ctx, `
      SELECT token, expires_at, created_at 
      FROM sessions  
      WHERE user_id = $1`,
      user.UserID)
    if err != nil {
      t.Fatalf("Error querying db for test results: %v", err)
    }
    assert.True(t, sessionR.Next())
    sessionR.Scan(&session.Token, &session.ExpiresAt, &session.CreatedAt)
    assert.NotEmpty(t, session.Token)
    assert.True(t, session.CreatedAt.Valid)
    assert.True(t, session.ExpiresAt.Valid)
    assert.Greater(t, session.ExpiresAt.Time, session.CreatedAt.Time)
    assert.True(t, recentTime.Before(session.CreatedAt.Time))
  })
}
