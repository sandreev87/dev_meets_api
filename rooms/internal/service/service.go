package service

import (
	"log/slog"
)

type Service struct {
	*RoomService
}

func NewService(logger *slog.Logger) *Service {
	return &Service{
		RoomService: NewRoomService(logger),
	}
}
