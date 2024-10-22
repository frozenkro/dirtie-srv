// Justification for factory pattern here is the shared
// connection pool between all the postgres repositories.
//
// The singleton instance of repoFactoryImpl is holding the
// pool and passes it to each repository.
package repos

import (
	"context"

	"github.com/frozenkro/dirtie-srv/internal/db"
)

type RepoFactory struct {
	tm *TxManager
}

func NewRepoFactory(ctx context.Context) (RepoFactory, error) {
	pool, err := db.PgConnect(ctx)
	if err != nil {
		return RepoFactory{}, err
	}

	tm := NewTxManager(pool)

	return RepoFactory{tm: tm}, nil
}

func (f RepoFactory) NewUserRepo() UserRepo {
	return UserRepo{sr: f.tm}
}

func (f RepoFactory) NewDeviceRepo() DeviceRepo {
	return DeviceRepo{sr: f.tm}
}

func (f RepoFactory) NewSessionRepo() SessionRepo {
	return SessionRepo{sr: f.tm}
}

func (f RepoFactory) NewProvisionStagingRepo() ProvisionStagingRepo {
	return ProvisionStagingRepo{sr: f.tm}
}

func (f RepoFactory) NewPwResetRepo() PwResetRepo {
	return PwResetRepo{sr: f.tm}
}
