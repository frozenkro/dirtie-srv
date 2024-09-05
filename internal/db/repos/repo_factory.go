// Justification for factory pattern here is the shared
// connection pool between all the postgres repositories.
//
// The singleton instance of repoFactoryImpl is holding the 
// pool and passes it to each repository.
package repos

import (
  "github.com/frozenkro/dirtie-srv/internal/db"
)

type RepoFactory interface {
  NewUserRepo() UserRepo
  NewDeviceRepo() DeviceRepo
  NewSessionRepo() SessionRepo
  NewProvisionStagingRepo() ProvisionStagingRepo
  NewPwResetRepo() PwResetRepo
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
  return &userRepoImpl{sr: f.tm}
}

func (f *repoFactoryImpl) NewDeviceRepo() DeviceRepo {
  return &deviceRepoImpl{sr: f.tm}
}

func (f *repoFactoryImpl) NewSessionRepo() SessionRepo {
  return &sessionRepoImpl{sr: f.tm}
}

func (f *repoFactoryImpl) NewProvisionStagingRepo() ProvisionStagingRepo {
  return &provisionStagingRepoImpl{sr: f.tm}
}

func (f *repoFactoryImpl) NewPwResetRepo() PwResetRepo {
  return &pwResetRepoImpl{sr: f.tm}
}
