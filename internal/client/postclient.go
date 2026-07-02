package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// PostClient resolves a post's author from post-service — like-service deliberately
// stores only (post_id, user_id) like rows and has no idea who authored a given post,
// so publishing post.liked (which needs the author's ID to notify them) requires this
// one lookup. TODO: replace with a gRPC client once post-service exposes one.
type PostClient struct {
	baseURL string
	http    *http.Client
}

func NewPostClient(baseURL string) *PostClient {
	return &PostClient{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

func (c *PostClient) GetPostAuthor(ctx context.Context, postID uuid.UUID) (uuid.UUID, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/posts/"+postID.String(), nil)
	if err != nil {
		return uuid.Nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return uuid.Nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, fmt.Errorf("post-service lookup failed: unexpected status %d", resp.StatusCode)
	}

	var post struct {
		UserID uuid.UUID `json:"user_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return uuid.Nil, err
	}
	return post.UserID, nil
}
