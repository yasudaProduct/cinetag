-- +goose Up
-- ================================================================
-- Baseline migration: cinetag 初期スキーマ
-- 既存DBとの互換性のため CREATE TABLE IF NOT EXISTS を使用
-- GORMモデル定義 (src/internal/model/*.go) の gorm タグから再現
-- ================================================================

CREATE TABLE IF NOT EXISTS users (
    id           uuid        NOT NULL DEFAULT gen_random_uuid(),
    clerk_user_id text       NOT NULL,
    display_id   text        NOT NULL,
    display_name text        NOT NULL,
    email        text        NOT NULL,
    avatar_url   text,
    bio          text,
    created_at   timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   timestamptz,

    CONSTRAINT users_pkey PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS users_clerk_user_id_key ON users (clerk_user_id);
CREATE UNIQUE INDEX IF NOT EXISTS users_display_id_key ON users (display_id);

CREATE TABLE IF NOT EXISTS tags (
    id               uuid        NOT NULL DEFAULT gen_random_uuid(),
    user_id          uuid        NOT NULL,
    title            text        NOT NULL,
    description      text,
    cover_image_url  text,
    is_public        boolean     NOT NULL DEFAULT false,
    add_movie_policy text        NOT NULL DEFAULT 'everyone',
    created_at       timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT tags_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS tag_movies (
    id               uuid        NOT NULL DEFAULT gen_random_uuid(),
    tag_id           uuid        NOT NULL,
    tmdb_movie_id    integer     NOT NULL,
    added_by_user_id uuid        NOT NULL,
    note             text,
    "position"       integer     NOT NULL DEFAULT 0,
    created_at       timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT tag_movies_pkey PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS tag_movies_unique ON tag_movies (tag_id, tmdb_movie_id);

CREATE TABLE IF NOT EXISTS tag_followers (
    tag_id     uuid        NOT NULL,
    user_id    uuid        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT tag_followers_pkey PRIMARY KEY (tag_id, user_id)
);

CREATE TABLE IF NOT EXISTS user_followers (
    follower_id uuid        NOT NULL,
    followee_id uuid        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT user_followers_pkey PRIMARY KEY (follower_id, followee_id)
);

CREATE TABLE IF NOT EXISTS movie_cache (
    tmdb_movie_id        integer     NOT NULL,
    title                text        NOT NULL,
    original_title       text,
    poster_path          text,
    backdrop_path        text,
    release_date         date,
    vote_average         numeric(3,1),
    overview             text,
    genres               jsonb,
    runtime              integer,
    production_countries jsonb,
    credits              jsonb,
    cached_at            timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at           timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP + interval '7 days'),

    CONSTRAINT movie_cache_pkey PRIMARY KEY (tmdb_movie_id)
);

-- +goose Down

DROP TABLE IF EXISTS movie_cache;
DROP TABLE IF EXISTS user_followers;
DROP TABLE IF EXISTS tag_followers;
DROP TABLE IF EXISTS tag_movies;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;
