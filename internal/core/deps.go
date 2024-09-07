package core

import (
  "github.com/frozenkro/dirtie-srv/internal/db/repos"
  "github.com/frozenkro/dirtie-srv/internal/services"
)

type Deps struct {
  AuthSvc services.AuthSvc
  DeviceRepo repos.DeviceRepo
  ProvStgRepo repos.ProvisionStagingRepo
  PwResetRepo repos.PwResetRepo
  SessionRepo repos.SessionRepo
  UserRepo repos.UserRepo
}

func NewDeps() *Deps {
  rf, err := repos.NewRepoFactory()
  if err != nil {
    panic("Failed to setup repositories")
  }

  deviceRepo := rf.NewDeviceRepo()
  provStgRepo := rf.NewProvisionStagingRepo()
  pwResetRepo := rf.NewPwResetRepo()
  sessionRepo := rf.NewSessionRepo()
  userRepo := rf.NewUserRepo()

  authSvc := services.NewAuthSvc(userRepo, sessionRepo)
  
  return &Deps{
    AuthSvc: *authSvc,
    DeviceRepo: deviceRepo,
    ProvStgRepo: provStgRepo,
    PwResetRepo: pwResetRepo,
    SessionRepo: sessionRepo,
    UserRepo: userRepo,
  }
}
