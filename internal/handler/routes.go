package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/snaply/like-service/internal/service"
	"go.uber.org/zap"
)

func NewRouter(likes service.LikeService, log *zap.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	likeH := NewLikeHandler(likes, log)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1/likes", func(r chi.Router) {
		r.Post("/batch", likeH.Batch)

		r.Route("/{post_id}", func(r chi.Router) {
			r.Post("/", likeH.Like)
			r.Delete("/", likeH.Unlike)
			r.Get("/likers", likeH.Likers)
		})
	})

	return r
}
