package repos

import (
  "context"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type DeviceRepo interface {
  CreateDevice(ctx context.Context, userId int32, displayName string) (sqlc.Device, error)
  GetDeviceByMacAddress(ctx context.Context, macAddr string) (sqlc.Device, error)
  GetDevicesByUser(ctx context.Context, userId int32) ([]sqlc.Device, error)
  RenameDevice(ctx context.Context, deviceId int32, displayName string) error
  UpdateDeviceMacAddress(ctx context.Context, deviceId int32, macAddr string) error
}

type deviceRepoImpl struct {
  sr SqlRunner
}

func (r *deviceRepoImpl) CreateDevice(ctx context.Context, userId int32, displayName string) (sqlc.Device, error) {
  res, err := r.sr.Query(ctx, func (q *sqlc.Queries) (interface{}, error) {
    params := sqlc.CreateDeviceParams {
      UserID: userId,
      DisplayName: pgtype.Text{ String: displayName },
    }
    return q.CreateDevice(ctx, params)
  })
  if err != nil || res == nil {
    return sqlc.Device{}, err
  }
  return res.(sqlc.Device), err
}

func (r *deviceRepoImpl) GetDeviceByMacAddress(ctx context.Context, macAddr string) (sqlc.Device, error) {
  res, err := r.sr.Query(ctx, func (q *sqlc.Queries) (interface{}, error) {
    return q.GetDeviceByMacAddress(ctx, pgtype.Text{String: macAddr})
  })
  return res.(sqlc.Device), err
}

func (r *deviceRepoImpl) GetDevicesByUser(ctx context.Context, userId int32) ([]sqlc.Device, error) {
  res, err := r.sr.Query(ctx, func (q *sqlc.Queries) (interface{}, error) {
    return q.GetDevicesByUser(ctx, userId)
  })
  return res.([]sqlc.Device), err
}

func (r *deviceRepoImpl) RenameDevice(ctx context.Context, deviceId int32, displayName string) error {
  return r.sr.Execute(ctx, func (q *sqlc.Queries) error {
    params := sqlc.RenameDeviceParams{
      DeviceID: deviceId,
      DisplayName: pgtype.Text{String: displayName},
    }
    return q.RenameDevice(ctx, params) 
  })
}

func (r *deviceRepoImpl) UpdateDeviceMacAddress(ctx context.Context, deviceId int32, macAddr string) error {
  return r.sr.Execute(ctx, func (q *sqlc.Queries) error {
    params := sqlc.UpdateDeviceMacAddressParams{
      DeviceID: deviceId,
      MacAddr: pgtype.Text{String: macAddr},
    }
    return q.UpdateDeviceMacAddress(ctx, params)
  })
}

