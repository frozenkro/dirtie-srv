package dto

import "github.com/frozenkro/dirtie-srv/internal/db/sqlc"

type DeviceDto struct {
	DeviceId    int32  `json:"deviceId"`
	UserId      int32  `json:"userId"`
	MacAddr     string `json:"macAddr"`
	DisplayName string `json:"displayName"`
}

func NewDeviceDto(d sqlc.Device) *DeviceDto {
	return &DeviceDto{
		DeviceId:    d.DeviceID,
		UserId:      d.UserID,
		MacAddr:     d.MacAddr.String,
		DisplayName: d.DisplayName.String,
	}
}
