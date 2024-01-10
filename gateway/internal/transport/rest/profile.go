package rest

import (
	"dev_meets/internal/transport"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strings"
)

type ProfileHandler struct {
	services transport.UserServiceInt
	logger   *slog.Logger
}

func NewProfileHandler(serv transport.UserServiceInt, logger *slog.Logger) *ProfileHandler {
	return &ProfileHandler{services: serv, logger: logger}
}

type ProfileResponse struct {
	Email string `json:"email" example:"email@gmail.com"`
}

type OkProfileResponse struct {
	Status  string          `json:"status" example:"ok"`
	Profile ProfileResponse `json:"profile"`
}

// Профиль текущего пользователя
// @Summary Профиль текущего пользователя
// @Tags Пользователь
// @Success 200 {object} OkProfileResponse "Параметры текущего пользователя"
// @Failure 201 {object} ErrResponse "Внутренняя ошибка сервиса"
// @Router /api/v1/personal-profile [get]
func (h *ProfileHandler) PersonalProfile(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.Replace(authorization, "Bearer", "", 1))
	user, err := h.services.CurrentUser(token)
	if err != nil {
		render.JSON(w, r, ErrResponse{
			Status: "internal_server_error",
		})
		return
	}

	response := OkProfileResponse{
		Status: "ok",
		Profile: ProfileResponse{
			Email: user.Email,
		},
	}
	render.JSON(w, r, response)
}
