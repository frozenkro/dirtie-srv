package repos

import (
	"context"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type DeviceRepo struct {
	sr SqlRunner
}

func (r DeviceRepo) CreateDevice(ctx context.Context, userId int32, displayName string) (sqlc.Device, error) {
	res, err := r.sr.Query(ctx, func(q *sqlc.Queries) (interface{}, error) {
		params := sqlc.CreateDeviceParams{
			UserID:      userId,
			DisplayName: pgtype.Text{String: displayName},
		}
		return q.CreateDevice(ctx, params)
	})
	if err != nil || res == nil {
		return sqlc.Device{}, err
	}
	return res.(sqlc.Device), err
}

func (r DeviceRepo) GetDeviceByMacAddress(ctx context.Context, macAddr string) (sqlc.Device, error) {
	res, err := r.sr.Query(ctx, func(q *sqlc.Queries) (interface{}, error) {
    return q.GetDeviceByMacAddress(ctx, pgtype.Text{String: macAddr, Valid: true})
	})

	if err != nil || res == nil {
		return sqlc.Device{}, err
	}
	return res.(sqlc.Device), err
}

func (r DeviceRepo) GetDevicesByUser(ctx context.Context, userId int32) ([]sqlc.Device, error) {
	res, err := r.sr.Query(ctx, func(q *sqlc.Queries) (interface{}, error) {
		return q.GetDevicesByUser(ctx, userId)
	})

	if err != nil || res == nil {
		return nil, err
	}
	return res.([]sqlc.Device), err
}

func (r DeviceRepo) RenameDevice(ctx context.Context, deviceId int32, displayName string) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		params := sqlc.RenameDeviceParams{
			DeviceID:    deviceId,
			DisplayName: pgtype.Text{String: displayName},
		}
		return q.RenameDevice(ctx, params)
	})
}

func (r DeviceRepo) UpdateDeviceMacAddress(ctx context.Context, deviceId int32, macAddr string) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		params := sqlc.UpdateDeviceMacAddressParams{
			DeviceID: deviceId,
			MacAddr:  pgtype.Text{String: macAddr},
		}
		return q.UpdateDeviceMacAddress(ctx, params)
	})
}
