package di

import (
  "context"

	"github.com/frozenkro/dirtie-srv/internal/core/utils"
	"github.com/frozenkro/dirtie-srv/internal/db"
	"github.com/frozenkro/dirtie-srv/internal/db/repos"
	brdcrm_topic "github.com/frozenkro/dirtie-srv/internal/hub/topics/brdcrmtopic"
	prv_topic "github.com/frozenkro/dirtie-srv/internal/hub/topics/prvtopic"
	"github.com/frozenkro/dirtie-srv/internal/services"
)

type Deps struct {
	BrdCrmTopic    *brdcrm_topic.BrdCrmTopic
	ProvisionTopic *prv_topic.ProvisionTopic

	AuthSvc   services.AuthSvc
	BrdCrmSvc services.BrdCrmSvc
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

	authSvc := services.NewAuthSvc(userRepo, userRepo, sessionRepo, pwResetRepo, htmlUtil, emailUtil)
	deviceSvc := services.NewDeviceSvc(deviceRepo, provStgRepo, ctxUtil)
	brdCrmSvc := services.NewBrdCrmSvc(influxRepo, influxRepo, deviceSvc)

	brdCrmTopic := brdcrm_topic.NewBrdCrmTopic(brdCrmSvc)
	prvTopic := prv_topic.NewProvisionTopic(*deviceSvc)

	return &Deps{
		BrdCrmTopic:         brdCrmTopic,
		ProvisionTopic:      prvTopic,
		AuthSvc:             *authSvc,
		BrdCrmSvc:           brdCrmSvc,
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
