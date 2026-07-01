CREATE TABLE likes (
    post_id     UUID        NOT NULL,
    user_id     UUID        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

CREATE INDEX idx_likes_post ON likes (post_id, created_at, user_id);
CREATE INDEX idx_likes_user ON likes (user_id, post_id);
