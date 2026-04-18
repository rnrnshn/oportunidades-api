-- Create "articles" table
CREATE TABLE "articles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "slug" text NOT NULL,
  "title" text NOT NULL,
  "excerpt" text NULL,
  "content" text NOT NULL,
  "cover_image_url" text NULL,
  "type" text NOT NULL,
  "status" text NOT NULL DEFAULT 'draft',
  "source_name" text NULL,
  "source_url" text NULL,
  "seo_title" text NULL,
  "seo_description" text NULL,
  "is_featured" boolean NOT NULL DEFAULT false,
  "author_id" uuid NOT NULL,
  "published_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "articles_slug_unique" UNIQUE ("slug"),
  CONSTRAINT "articles_author_id_fkey" FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "articles_status_check" CHECK (status = ANY (ARRAY['draft'::text, 'in_review'::text, 'published'::text, 'archived'::text])),
  CONSTRAINT "articles_type_check" CHECK (type = ANY (ARRAY['editorial'::text, 'news'::text, 'guide'::text]))
);
-- Create index "articles_author_id_idx" to table: "articles"
CREATE INDEX "articles_author_id_idx" ON "articles" ("author_id") WHERE (deleted_at IS NULL);
-- Create index "articles_deleted_at_idx" to table: "articles"
CREATE INDEX "articles_deleted_at_idx" ON "articles" ("deleted_at");
-- Create index "articles_is_featured_idx" to table: "articles"
CREATE INDEX "articles_is_featured_idx" ON "articles" ("is_featured") WHERE (deleted_at IS NULL);
-- Create index "articles_published_at_idx" to table: "articles"
CREATE INDEX "articles_published_at_idx" ON "articles" ("published_at") WHERE (deleted_at IS NULL);
-- Create index "articles_status_idx" to table: "articles"
CREATE INDEX "articles_status_idx" ON "articles" ("status") WHERE (deleted_at IS NULL);
-- Create index "articles_type_idx" to table: "articles"
CREATE INDEX "articles_type_idx" ON "articles" ("type") WHERE (deleted_at IS NULL);
-- Create trigger "articles_set_updated_at"
CREATE TRIGGER "articles_set_updated_at" BEFORE UPDATE ON "articles" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
