package di

import (
	"context"

	"github.com/frozenkro/dirtie-srv/internal/core/utils"
	"github.com/frozenkro/dirtie-srv/internal/db"
	"github.com/frozenkro/dirtie-srv/internal/db/repos"
	"github.com/frozenkro/dirtie-srv/internal/hub/topics/brdcrmtopic"
	"github.com/frozenkro/dirtie-srv/internal/hub/topics/logdumptopic"
	"github.com/frozenkro/dirtie-srv/internal/hub/topics/prvtopic"
	"github.com/frozenkro/dirtie-srv/internal/services"
)

type Deps struct {
	BrdCrmTopic    *brdcrmtopic.BrdCrmTopic
	LogDumpTopic   *logdumptopic.LogDumpTopic
	ProvisionTopic *prvtopic.ProvisionTopic

	AuthSvc    services.AuthSvc
	BrdCrmSvc  services.BrdCrmSvc
	DataSvc    services.DataSvc
	DeviceSvc  services.DeviceSvc
	LogDumpSvc services.LogDumpSvc

	DeviceRepo  repos.DeviceRepo
	ProvStgRepo repos.ProvisionStagingRepo
	PwResetRepo repos.PwResetRepo
	SessionRepo repos.SessionRepo
	UserRepo    repos.UserRepo

	InfluxRepo db.InfluxRepo

	EmailUtil utils.EmailUtil
	HtmlUtil  utils.HtmlUtil
	CtxUtil   utils.CtxUtil
}

// context is just used for passing test-specific config around
// Main app should just pass background context
func NewDeps(ctx context.Context) *Deps {
	rf, err := repos.NewRepoFactory(ctx)
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

	authSvc := services.NewAuthSvc(userRepo,
		userRepo,
		sessionRepo,
		sessionRepo,
		pwResetRepo,
		pwResetRepo,
		htmlUtil,
		emailUtil)
	deviceSvc := services.NewDeviceSvc(deviceRepo,
		deviceRepo,
		provStgRepo,
		provStgRepo,
		ctxUtil)
	brdCrmSvc := services.NewBrdCrmSvc(
		influxRepo,
		influxRepo,
		deviceSvc,
		deviceSvc,
	)
	logDumpSvc := services.NewLogDumpSvc(
		deviceSvc,
	)
	dataSvc := services.NewDataSvc(
		influxRepo)

	brdCrmTopic := brdcrmtopic.NewBrdCrmTopic(brdCrmSvc)
	logDumpTopic := logdumptopic.NewLogDumpTopic(logDumpSvc)
	prvTopic := prvtopic.NewProvisionTopic(*deviceSvc)

	return &Deps{
		BrdCrmTopic:    brdCrmTopic,
		LogDumpTopic:   logDumpTopic,
		ProvisionTopic: prvTopic,
		AuthSvc:        authSvc,
		BrdCrmSvc:      brdCrmSvc,
		DataSvc:        dataSvc,
		DeviceSvc:      *deviceSvc,
		DeviceRepo:     deviceRepo,
		ProvStgRepo:    provStgRepo,
		PwResetRepo:    pwResetRepo,
		SessionRepo:    sessionRepo,
		UserRepo:       userRepo,
		EmailUtil:      *emailUtil,
		HtmlUtil:       *htmlUtil,
		CtxUtil:        *ctxUtil,
		InfluxRepo:     influxRepo,
	}
}
