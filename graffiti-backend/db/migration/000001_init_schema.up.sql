CREATE TYPE "status" AS ENUM ('pending', 'friends', 'blocked');

CREATE TYPE "post_type" AS ENUM ('media', 'embed_link');

CREATE TABLE
  "users" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    "username" varchar UNIQUE NOT NULL,
    "fullname" varchar,
    "email" varchar UNIQUE NOT NULL,
    "hashed_password" varchar NOT NULL,
    "profile_picture" varchar,
    "bio" varchar,
    "has_onboarded" boolean DEFAULT false,
    "background_image" varchar,
    "onboarding_at" timestamp,
    "created_at" timestamp DEFAULT now (),
    "updated_at" timestamp DEFAULT now ()
  );

CREATE TABLE
  "walls" (
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid ()),
    "user_id" uuid NOT NULL,
    "title" varchar NOT NULL,
    "description" varchar,
    "background_image" varchar,
    "is_public" boolean DEFAULT false,
    "is_archived" boolean DEFAULT false,
    "is_deleted" boolean DEFAULT false,
    "popularity_score" float DEFAULT 0,
    "created_at" timestamp DEFAULT (now ()),
    "updated_at" timestamp DEFAULT (now ())
  );

CREATE TABLE
  "posts" (
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid ()),
    "wall_id" uuid,
    "author" uuid NOT NULL,
    "media_url" varchar,
    "post_type" post_type,
    "is_highlighted" boolean DEFAULT false,
    "likes_count" integer DEFAULT 0,
    "is_deleted" boolean DEFAULT false,
    "created_at" timestamp DEFAULT (now ())
  );

CREATE TABLE
  "likes" (
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid ()),
    "post_id" uuid NOT NULL,
    "user_id" uuid NOT NULL,
    "liked_at" timestamp DEFAULT (now ())
  );

CREATE TABLE
  "friendships" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    "from_user" uuid NOT NULL,
    "to_user" uuid NOT NULL,
    "status" status,
    "created_at" timestamp DEFAULT now (),
    "updated_at" timestamp DEFAULT now (),
    CONSTRAINT "friendships_from_user_fk" FOREIGN KEY ("from_user") REFERENCES "users" ("id") ON DELETE CASCADE,
    CONSTRAINT "friendships_to_user_fk" FOREIGN KEY ("to_user") REFERENCES "users" ("id") ON DELETE CASCADE,
    CONSTRAINT "unique_friendship" UNIQUE ("from_user", "to_user")
  );

-- Materialized view for accepted (mutual) friendships
CREATE MATERIALIZED VIEW accepted_friendships_mv AS
SELECT DISTINCT
    from_user AS user_id,
    to_user AS friend_id
FROM friendships
WHERE status = 'friends'
UNION
SELECT DISTINCT
    to_user AS user_id,
    from_user AS friend_id
FROM friendships
WHERE status = 'friends';

CREATE INDEX ON "users" ("username");

CREATE INDEX ON "walls" ("user_id");

CREATE INDEX ON "likes" ("post_id");

CREATE INDEX ON "likes" ("user_id");

CREATE INDEX ON "likes" ("post_id", "user_id");

CREATE INDEX ON "friendships" ("from_user");

CREATE INDEX ON "friendships" ("to_user");

CREATE INDEX ON "friendships" ("from_user", "to_user");

-- Use trigrams for fuzzy search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_users_username_trgm ON users USING gin (username gin_trgm_ops);
CREATE INDEX idx_users_fullname_trgm ON users USING gin (fullname gin_trgm_ops);

ALTER DATABASE graffiti SET pg_trgm.similarity_threshold = 0;

-- Indexes for materialized view
CREATE INDEX idx_accepted_friendships_user_id ON accepted_friendships_mv (user_id);
CREATE INDEX idx_accepted_friendships_friend_id ON accepted_friendships_mv (friend_id);
CREATE UNIQUE INDEX idx_accepted_friendships_mv_unique ON accepted_friendships_mv (user_id, friend_id);

