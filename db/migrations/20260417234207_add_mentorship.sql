-- Create "mentor_profiles" table
CREATE TABLE "mentor_profiles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "headline" text NOT NULL,
  "bio" text NOT NULL,
  "expertise" text NOT NULL,
  "availability" text NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "mentor_profiles_user_id_unique" UNIQUE ("user_id"),
  CONSTRAINT "mentor_profiles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "mentor_profiles_deleted_at_idx" to table: "mentor_profiles"
CREATE INDEX "mentor_profiles_deleted_at_idx" ON "mentor_profiles" ("deleted_at");
-- Create index "mentor_profiles_is_active_idx" to table: "mentor_profiles"
CREATE INDEX "mentor_profiles_is_active_idx" ON "mentor_profiles" ("is_active") WHERE (deleted_at IS NULL);
-- Create index "mentor_profiles_user_id_idx" to table: "mentor_profiles"
CREATE INDEX "mentor_profiles_user_id_idx" ON "mentor_profiles" ("user_id") WHERE (deleted_at IS NULL);
-- Create trigger "mentor_profiles_set_updated_at"
CREATE TRIGGER "mentor_profiles_set_updated_at" BEFORE UPDATE ON "mentor_profiles" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create "mentorship_sessions" table
CREATE TABLE "mentorship_sessions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "mentor_id" uuid NOT NULL,
  "requester_id" uuid NOT NULL,
  "message" text NOT NULL,
  "status" text NOT NULL DEFAULT 'pending',
  "scheduled_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "mentorship_sessions_mentor_id_fkey" FOREIGN KEY ("mentor_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "mentorship_sessions_requester_id_fkey" FOREIGN KEY ("requester_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "mentorship_sessions_status_check" CHECK (status = ANY (ARRAY['pending'::text, 'accepted'::text, 'rejected'::text, 'completed'::text, 'cancelled'::text]))
);
-- Create index "mentorship_sessions_deleted_at_idx" to table: "mentorship_sessions"
CREATE INDEX "mentorship_sessions_deleted_at_idx" ON "mentorship_sessions" ("deleted_at");
-- Create index "mentorship_sessions_mentor_id_idx" to table: "mentorship_sessions"
CREATE INDEX "mentorship_sessions_mentor_id_idx" ON "mentorship_sessions" ("mentor_id") WHERE (deleted_at IS NULL);
-- Create index "mentorship_sessions_requester_id_idx" to table: "mentorship_sessions"
CREATE INDEX "mentorship_sessions_requester_id_idx" ON "mentorship_sessions" ("requester_id") WHERE (deleted_at IS NULL);
-- Create index "mentorship_sessions_status_idx" to table: "mentorship_sessions"
CREATE INDEX "mentorship_sessions_status_idx" ON "mentorship_sessions" ("status") WHERE (deleted_at IS NULL);
-- Create trigger "mentorship_sessions_set_updated_at"
CREATE TRIGGER "mentorship_sessions_set_updated_at" BEFORE UPDATE ON "mentorship_sessions" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
