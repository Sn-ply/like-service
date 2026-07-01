package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/snaply/like-service/internal/service"
	"go.uber.org/zap"
)

type LikeHandler struct {
	likes service.LikeService
	log   *zap.Logger
}

func NewLikeHandler(likes service.LikeService, log *zap.Logger) *LikeHandler {
	return &LikeHandler{likes: likes, log: log}
}

func targetPostID(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "post_id"))
	return id, err == nil
}

func (h *LikeHandler) Like(w http.ResponseWriter, r *http.Request) {
	userID, ok := callerID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "missing user identity")
		return
	}
	postID, ok := targetPostID(r)
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid post_id")
		return
	}

	if err := h.likes.Like(r.Context(), postID, userID); err != nil {
		h.log.Error("like error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LikeHandler) Unlike(w http.ResponseWriter, r *http.Request) {
	userID, ok := callerID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "missing user identity")
		return
	}
	postID, ok := targetPostID(r)
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid post_id")
		return
	}

	if err := h.likes.Unlike(r.Context(), postID, userID); err != nil {
		h.log.Error("unlike error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LikeHandler) Likers(w http.ResponseWriter, r *http.Request) {
	postID, ok := targetPostID(r)
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid post_id")
		return
	}

	page, err := h.likes.Likers(r.Context(), postID, r.URL.Query().Get("cursor"), parseLimit(r, 20))
	if err != nil {
		h.log.Error("likers error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"data":        page.UserIDs,
		"next_cursor": page.NextCursor,
	})
}

func (h *LikeHandler) Batch(w http.ResponseWriter, r *http.Request) {
	userID, ok := callerID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	var req struct {
		PostIDs []string `json:"post_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.PostIDs) > 100 {
		respondError(w, http.StatusBadRequest, "too many post_ids (max 100)")
		return
	}

	postIDs := make([]uuid.UUID, 0, len(req.PostIDs))
	for _, s := range req.PostIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid post_id: "+s)
			return
		}
		postIDs = append(postIDs, id)
	}

	summaries, err := h.likes.Batch(r.Context(), postIDs, userID)
	if err != nil {
		h.log.Error("batch error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, summaries)
}
