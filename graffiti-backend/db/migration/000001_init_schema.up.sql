CREATE TYPE "status" AS ENUM (
  'pending',
  'friends',
  'blocked'
);

CREATE TYPE "post_type" AS ENUM (
  'image',
  'video',
  'text',
  'gif'
);

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY,
  "username" varchar UNIQUE NOT NULL,
  "fullname" varchar,
  "email" varchar NOT NULL,
  "hashed_password" varchar NOT NULL,
  "profile_picture" varchar,
  "bio" varchar,
  "has_onboarded" boolean DEFAULT false,
  "background_image" varchar,
  "onboarding_at" timestamp,
  "created_at" timestamp DEFAULT now(),
  "updated_at" timestamp DEFAULT now()
);

CREATE TABLE "walls" (
  "id" uuid PRIMARY KEY,
  "user_id" uuid NOT NULL,
  "description" varchar,
  "background_image" varchar,
  "is_public" boolean DEFAULT true,
  "is_archived" boolean DEFAULT false,
  "is_deleted" boolean DEFAULT false,
  "popularity_score" float DEFAULT 0,
  "created_at" timestamp DEFAULT now(),
  "updated_at" timestamp DEFAULT now(),
  CONSTRAINT "wall_users" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE TABLE "posts" (
  "id" uuid PRIMARY KEY,
  "wall_id" uuid,
  "author" uuid NOT NULL, -- ✅ Changed from varchar to uuid
  "media_url" varchar,
  "post_type" post_type,
  "is_highlighted" boolean DEFAULT false,
  "likes_count" integer DEFAULT 0,
  "is_deleted" boolean DEFAULT false,
  "created_at" timestamp DEFAULT now(),
  CONSTRAINT "wall_posts" FOREIGN KEY ("wall_id") REFERENCES "walls" ("id") ON DELETE CASCADE,
  CONSTRAINT "post_author_fk" FOREIGN KEY ("author") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE TABLE "likes" (
  "id" uuid PRIMARY KEY,
  "post_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "liked_at" timestamp DEFAULT now(),
  CONSTRAINT "post_likes" FOREIGN KEY ("post_id") REFERENCES "posts" ("id") ON DELETE CASCADE,
  CONSTRAINT "user_likes" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE TABLE "friendships" (
  "id" uuid PRIMARY KEY,
  "user_id" uuid NOT NULL,
  "friend_id" uuid NOT NULL,
  "status" status,
  "created_at" timestamp DEFAULT now(),
  "updated_at" timestamp DEFAULT now(),
  CONSTRAINT "friendships_user_fk" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
  CONSTRAINT "friendships_friend_fk" FOREIGN KEY ("friend_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

-- ✅ INDEXES
CREATE INDEX ON "users" ("username");

CREATE INDEX ON "walls" ("user_id");

CREATE INDEX ON "likes" ("post_id");
CREATE INDEX ON "likes" ("user_id");
CREATE INDEX ON "likes" ("post_id", "user_id");

CREATE INDEX ON "friendships" ("user_id");
CREATE INDEX ON "friendships" ("friend_id");
CREATE INDEX ON "friendships" ("user_id", "friend_id");
