package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "SubServices/cmd/webserver/docs"
	"SubServices/internal/http/handlers"
)

func InitRouter(h *handlers.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", handlers.Health)
		r.Route("/subscriptions", func(r chi.Router) {
			r.Post("/", h.CreateSubscription)
			r.Get("/{id}", h.GetSubscription)
			r.Put("/{id}", h.UpdateSubscription)
			r.Delete("/{id}", h.DeleteSubscription)
			r.Get("/", h.ListSubscriptions)
		})
	})

	return r
}
