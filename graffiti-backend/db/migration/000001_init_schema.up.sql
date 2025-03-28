CREATE TYPE "status" AS ENUM ('pending', 'friends', 'blocked');

CREATE TYPE "post_type" AS ENUM ('media', 'embed_link');

CREATE TABLE
  "users" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
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
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id" uuid NOT NULL,
    "description" varchar,
    "background_image" varchar,
    "is_public" boolean DEFAULT false,
    "is_archived" boolean DEFAULT false,
    "is_deleted" boolean DEFAULT false,
    "popularity_score" float DEFAULT 0,
    "created_at" timestamp DEFAULT now (),
    "updated_at" timestamp DEFAULT now (),
    CONSTRAINT "wall_users" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
  );

CREATE TABLE
  "posts" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "wall_id" uuid,
    "author" uuid NOT NULL, -- ✅ Changed from varchar to uuid
    "media_url" varchar,
    "post_type" post_type,
    "is_highlighted" boolean DEFAULT false,
    "likes_count" integer DEFAULT 0,
    "is_deleted" boolean DEFAULT false,
    "created_at" timestamp DEFAULT now (),
    CONSTRAINT "wall_posts" FOREIGN KEY ("wall_id") REFERENCES "walls" ("id") ON DELETE CASCADE,
    CONSTRAINT "post_author_fk" FOREIGN KEY ("author") REFERENCES "users" ("id") ON DELETE CASCADE
  );

CREATE TABLE
  "likes" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "post_id" uuid NOT NULL,
    "user_id" uuid NOT NULL,
    "liked_at" timestamp DEFAULT now (),
    CONSTRAINT "post_likes" FOREIGN KEY ("post_id") REFERENCES "posts" ("id") ON DELETE CASCADE,
    CONSTRAINT "user_likes" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
  );

CREATE TABLE
  "friendships" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "from_user" uuid NOT NULL,
    "to_user" uuid NOT NULL,
    "status" status,
    "created_at" timestamp DEFAULT now (),
    "updated_at" timestamp DEFAULT now (),
    CONSTRAINT "friendships_from_user_fk" FOREIGN KEY ("from_user") REFERENCES "users" ("id") ON DELETE CASCADE,
    CONSTRAINT "friendships_to_user_fk" FOREIGN KEY ("to_user") REFERENCES "users" ("id") ON DELETE CASCADE,
    CONSTRAINT "unique_friendship" UNIQUE ("from_user", "to_user")
  );

-- ✅ INDEXES
CREATE INDEX ON "users" ("username");

CREATE INDEX ON "walls" ("user_id");

CREATE INDEX ON "likes" ("post_id");

CREATE INDEX ON "likes" ("user_id");

CREATE INDEX ON "likes" ("post_id", "user_id");

CREATE INDEX ON "friendships" ("from_user");

CREATE INDEX ON "friendships" ("to_user");

CREATE INDEX ON "friendships" ("from_user", "to_user");