package rest

import (
	_ "auth/api"
	"auth/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
)

type AuthorizationHandlerInt interface {
	signUp(w http.ResponseWriter, r *http.Request)
	signIn(w http.ResponseWriter, r *http.Request)
	logout(w http.ResponseWriter, r *http.Request)
	userIdentity(next http.Handler) http.Handler
}

type ProfileHandlerInt interface {
	PersonalProfile(w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	AuthorizationHandlerInt
	ProfileHandlerInt
}

func NewHandler(services *service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		AuthorizationHandlerInt: NewAuthHandler(services.AuthService, logger),
		ProfileHandlerInt:       NewProfileHandler(services.UserService, logger),
	}
}

func (h *Handler) InitRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)

	router.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("pong"))
			})

			r.Post("/sign-up", h.AuthorizationHandlerInt.signUp)
			r.Post("/sign-in", h.AuthorizationHandlerInt.signIn)
			r.Delete("/logout", h.AuthorizationHandlerInt.logout)

			r.Route("/personal-profile", func(r chi.Router) {
				r.Use(h.AuthorizationHandlerInt.userIdentity)
				r.Get("/", h.ProfileHandlerInt.PersonalProfile)
			})
		})
	})

	return router
}
