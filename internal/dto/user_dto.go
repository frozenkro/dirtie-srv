package dto

import "github.com/frozenkro/dirtie-srv/internal/db/sqlc"

type UserDto struct {
	UserId int32  `json:"userId"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func NewUserDto(u sqlc.User) *UserDto {
	return &UserDto{
		UserId: u.UserID,
		Email:  u.Email,
		Name:   u.Name,
	}
}
