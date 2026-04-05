-- +goose Up

CREATE TABLE IF NOT EXISTS tag_likes (
    tag_id     uuid        NOT NULL,
    user_id    uuid        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT tag_likes_pkey PRIMARY KEY (tag_id, user_id)
);

-- +goose Down

DROP TABLE IF EXISTS tag_likes;
