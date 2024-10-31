package repos

import (
	"context"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProvisionStagingRepo struct {
	sr SqlRunner
}

func (r ProvisionStagingRepo) CreateProvisionStaging(ctx context.Context, deviceId int32, contract string) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		params := sqlc.CreateProvisionStagingParams{
			DeviceID: deviceId,
			Contract: pgtype.Text{
				String: contract,
				Valid:  true,
			},
		}
		return q.CreateProvisionStaging(ctx, params)
	})
}

func (r ProvisionStagingRepo) GetProvisionStagingByContract(ctx context.Context, contract string) (sqlc.ProvisionStaging, error) {
	res, err := r.sr.Query(ctx, func(q *sqlc.Queries) (interface{}, error) {
		return q.GetProvisionStagingByContract(
			ctx,
			pgtype.Text{
				String: contract,
				Valid:  true,
			})
	})

	if err != nil || res == nil {
		return sqlc.ProvisionStaging{}, err
	}
	return res.(sqlc.ProvisionStaging), err
}

func (r ProvisionStagingRepo) DeleteProvisionStaging(ctx context.Context, deviceId int32) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		return q.DeleteProvisionStaging(ctx, deviceId)
	})
}
