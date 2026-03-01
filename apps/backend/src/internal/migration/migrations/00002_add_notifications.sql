-- +goose Up
-- ================================================================
-- 通知テーブル追加
-- Phase 1: アプリ内通知機能
-- ================================================================

CREATE TABLE notifications (
    id                UUID        NOT NULL DEFAULT gen_random_uuid(),
    recipient_user_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_user_id     UUID                 REFERENCES users(id) ON DELETE SET NULL,
    notification_type TEXT        NOT NULL,
    tag_id            UUID                 REFERENCES tags(id) ON DELETE CASCADE,
    tag_movie_id      UUID                 REFERENCES tag_movies(id) ON DELETE CASCADE,
    is_read           BOOLEAN     NOT NULL DEFAULT false,
    read_at           TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT notifications_pkey PRIMARY KEY (id)
);

CREATE INDEX idx_notifications_recipient_created
    ON notifications (recipient_user_id, created_at DESC);

CREATE INDEX idx_notifications_recipient_unread
    ON notifications (recipient_user_id, is_read)
    WHERE is_read = false;

-- +goose Down

DROP TABLE IF EXISTS notifications;
