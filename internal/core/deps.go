package core

import (
	"github.com/frozenkro/dirtie-srv/internal/db/repos"
	"github.com/frozenkro/dirtie-srv/internal/services"
  "github.com/frozenkro/dirtie-srv/internal/core/utils"
)

type Deps struct {
	AuthSvc   services.AuthSvc
	DeviceSvc services.DeviceSvc

	DeviceRepo  repos.DeviceRepo
	ProvStgRepo repos.ProvisionStagingRepo
	PwResetRepo repos.PwResetRepo
	SessionRepo repos.SessionRepo
	UserRepo    repos.UserRepo

  EmailSender utils.EmailSender
  HtmlParser  utils.HtmlParser
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

  emailUtil := &utils.EmailUtil{}
  htmlUtil := &utils.HtmlUtil{}

	authSvc := services.NewAuthSvc(userRepo, sessionRepo, pwResetRepo, htmlUtil, emailUtil)
	deviceSvc := services.NewDeviceSvc(deviceRepo, provStgRepo)

	return &Deps{
		AuthSvc:     *authSvc,
		DeviceSvc:   *deviceSvc,
		DeviceRepo:  deviceRepo,
		ProvStgRepo: provStgRepo,
		PwResetRepo: pwResetRepo,
		SessionRepo: sessionRepo,
		UserRepo:    userRepo,
    EmailSender: emailUtil,
    HtmlParser: htmlUtil,
	}
}
