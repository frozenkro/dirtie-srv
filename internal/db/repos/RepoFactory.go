package repos

import (
  "github.com/frozenkro/dirtie-srv/internal/db"
)

type RepoFactory interface {
  NewUserRepo() UserRepo
  NewDeviceRepo() DeviceRepo
  NewSessionRepo() SessionRepo
  NewProvisionStagingRepo() ProvisionStagingRepo
}

type repoFactoryImpl struct {
  tm *TxManager
}

func NewRepoFactory() (RepoFactory, error) {
  pool, err := db.PgConnect()
  if err != nil {
    return nil, err
  }

  tm := NewTxManager(pool)

  return &repoFactoryImpl{tm: tm}, nil
}

func (f *repoFactoryImpl) NewUserRepo() UserRepo {
  return &userRepoImpl{tm: f.tm}
}

func (f *repoFactoryImpl) NewDeviceRepo() DeviceRepo {
  return &deviceRepoImpl{tm: f.tm}
}

func (f *repoFactoryImpl) NewSessionRepo() SessionRepo {
  return &sessionRepoImpl{tm: f.tm}
}

func (f *repoFactoryImpl) NewProvisionStagingRepo() ProvisionStagingRepo {
  return &provisionStagingRepoImpl{tm: f.tm}
}
