package rest

import (
	"auth/internal/domain/models"
	"auth/internal/transport"
	"errors"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type AuthHandler struct {
	services transport.AuthorizationServiceInt
	logger   *slog.Logger
}

func NewAuthHandler(serv transport.AuthorizationServiceInt, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{services: serv, logger: logger}
}

type SignInOkResponse struct {
	Status string `json:"status" example:"ok"`
	Token  string `json:"token"  example:"adsghjyjh5effa234353ty..."`
}

type SignUpOkResponse struct {
	Status string `json:"status" example:"ok"`
	Id     int    `json:"id" example:"123"`
}

// Регистрация
// @Summary Регистрация нового пользователя
// @Tags Регистрация
// @Param Request body signInUpInput true "Почта и Пароль"
// @Success 200 {object} SignUpOkResponse "Успешная регистрация нового пользователя"
// @Failure 201 {object} ErrResponse "Ошибка при попытке зарегистрироваться"
// @Router /api/v1/sign-up [post]
func (h *AuthHandler) signUp(w http.ResponseWriter, r *http.Request) {
	var input signInUpInput

	err := render.DecodeJSON(r.Body, &input)
	if errors.Is(err, io.EOF) {
		// Такую ошибку встретим, если получили запрос с пустым телом.
		// Обработаем её отдельно
		h.logger.Error("request body is empty")

		render.JSON(w, r, ErrResponse{
			Status: "wrong_params",
		})

		return
	}
	if err != nil {
		h.logger.Error("failed to decode request body")

		render.JSON(w, r, ErrResponse{
			Status: "wrong_params",
		})

		return
	}

	h.logger.Info("request body decoded", slog.Any("request", input))

	if err := validator.New().Struct(input); err != nil {
		render.JSON(w, r, ErrResponse{
			Status: "wrong_params",
		})

		return
	}

	user := models.User{Email: input.Email}

	id, err := h.services.RegisterNewUser(user, input.Password)
	if err != nil {
		e := slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		}
		h.logger.Error("internal_server_error", e)
		render.JSON(w, r, ErrResponse{
			Status: "internal_server_error",
		})
		return
	}

	render.JSON(w, r, SignUpOkResponse{
		Status: "ok",
		Id:     id,
	})
}

type signInUpInput struct {
	Email    string `json:"email" validate:"required,email" example:"email@gmail.com"`
	Password string `json:"password" validate:"required" example:"password"`
}

// Авторизация
// @Summary Авторизация пользователя
// @Tags Авторизация
// @Param Request body signInUpInput true "Почта и Пароль"
// @Success 200 {object} SignInOkResponse "Успешная авторизация"
// @Failure 201 {object} ErrResponse "Ошибка при попытке авторизоваться"
// @Router /api/v1/sign-in [post]
func (h *AuthHandler) signIn(w http.ResponseWriter, r *http.Request) {
	var input signInUpInput

	err := render.DecodeJSON(r.Body, &input)
	if errors.Is(err, io.EOF) {
		// Такую ошибку встретим, если получили запрос с пустым телом.
		// Обработаем её отдельно
		h.logger.Error("request body is empty")

		render.JSON(w, r, ErrResponse{
			Status: "wrong_params",
		})

		return
	}
	if err != nil {
		h.logger.Error("failed to decode request body")

		render.JSON(w, r, ErrResponse{
			Status: "wrong_params",
		})

		return
	}

	h.logger.Info("request body decoded", slog.Any("request", input))

	if err := validator.New().Struct(input); err != nil {
		h.logger.Error("invalid params", err)
		render.JSON(w, r, ErrResponse{
			Status: "wrong_params",
		})

		return
	}

	token, err := h.services.Login(input.Email, input.Password)
	if err != nil {
		h.logger.Error("internal error", err)
		render.JSON(w, r, ErrResponse{
			Status: "internal_server_error",
		})
		return
	}

	render.JSON(w, r, SignInOkResponse{
		Status: "ok",
		Token:  token,
	})
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	if r.Header["Authorization"] == nil {
		h.logger.Error("invalid params")
		render.JSON(w, r, ErrResponse{
			Status: "wrong_params",
		})
		return
	}

	authorization := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.Replace(authorization, "Bearer", "", 1))
	if err := h.services.Logout(token); err != nil {
		h.logger.Error("internal error", err)
		render.JSON(w, r, ErrResponse{
			Status: "internal_server_error",
		})
		return
	}
	render.JSON(w, r, OkResponse{
		Status: "ok",
	})
}

func (h *AuthHandler) userIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Authorization"] == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		authorization := r.Header.Get("Authorization")
		token := strings.TrimSpace(strings.Replace(authorization, "Bearer", "", 1))
		if err := h.services.Verify(token); err != nil {
			e := slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			}
			h.logger.Error("failed to verify token", e)
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
