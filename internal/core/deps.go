package core

import (
	"github.com/frozenkro/dirtie-srv/internal/core/utils"
	"github.com/frozenkro/dirtie-srv/internal/db"
	"github.com/frozenkro/dirtie-srv/internal/db/repos"
	"github.com/frozenkro/dirtie-srv/internal/services"
)

type Deps struct {
	AuthSvc   services.AuthSvc
	DeviceSvc services.DeviceSvc

	DeviceRepo  repos.DeviceRepo
	ProvStgRepo repos.ProvisionStagingRepo
	PwResetRepo repos.PwResetRepo
	SessionRepo repos.SessionRepo
	UserRepo    repos.UserRepo

	DeviceDataRecorder  db.DeviceDataRecorder
	DeviceDataRetriever db.DeviceDataRetriever

	EmailSender utils.EmailSender
	HtmlParser  utils.HtmlParser
	UserGetter  utils.UserGetter
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

	influxRepo := db.NewInfluxRepo()

	emailUtil := &utils.EmailUtil{}
	htmlUtil := &utils.HtmlUtil{}
	ctxUtil := &utils.CtxUtil{}

	authSvc := services.NewAuthSvc(userRepo, sessionRepo, pwResetRepo, htmlUtil, emailUtil)
	deviceSvc := services.NewDeviceSvc(deviceRepo, provStgRepo, ctxUtil)

	return &Deps{
		AuthSvc:             *authSvc,
		DeviceSvc:           *deviceSvc,
		DeviceRepo:          deviceRepo,
		ProvStgRepo:         provStgRepo,
		PwResetRepo:         pwResetRepo,
		SessionRepo:         sessionRepo,
		UserRepo:            userRepo,
		EmailSender:         emailUtil,
		HtmlParser:          htmlUtil,
		DeviceDataRecorder:  influxRepo,
		DeviceDataRetriever: influxRepo,
	}
}
