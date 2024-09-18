package repos

import (
	"context"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProvisionStagingRepo interface {
	CreateProvisionStaging(ctx context.Context, deviceId int32, contract string) error
	GetProvisionStagingByContract(ctx context.Context, contract string) (sqlc.ProvisionStaging, error)
	DeleteProvisionStaging(ctx context.Context, deviceId int32) error
}

type provisionStagingRepoImpl struct {
	sr SqlRunner
}

func (r *provisionStagingRepoImpl) CreateProvisionStaging(ctx context.Context, deviceId int32, contract string) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		params := sqlc.CreateProvisionStagingParams{
			DeviceID: deviceId,
			Contract: pgtype.Text{String: contract},
		}
		return q.CreateProvisionStaging(ctx, params)
	})
}

func (r *provisionStagingRepoImpl) GetProvisionStagingByContract(ctx context.Context, contract string) (sqlc.ProvisionStaging, error) {
	res, err := r.sr.Query(ctx, func(q *sqlc.Queries) (interface{}, error) {
		return q.GetProvisionStagingByContract(ctx, pgtype.Text{String: contract})
	})

	if err != nil || res == nil {
		return sqlc.ProvisionStaging{}, err
	}
	return res.(sqlc.ProvisionStaging), err
}

func (r *provisionStagingRepoImpl) DeleteProvisionStaging(ctx context.Context, deviceId int32) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		return q.DeleteProvisionStaging(ctx, deviceId)
	})
}
