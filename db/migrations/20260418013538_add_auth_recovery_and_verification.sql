-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "email_verified_at" timestamptz NULL;
-- Create "auth_action_tokens" table
CREATE TABLE "auth_action_tokens" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "purpose" text NOT NULL,
  "token_hash" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "consumed_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "auth_action_tokens_token_hash_unique" UNIQUE ("token_hash"),
  CONSTRAINT "auth_action_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "auth_action_tokens_purpose_check" CHECK (purpose = ANY (ARRAY['password_reset'::text, 'email_verification'::text]))
);
-- Create index "auth_action_tokens_deleted_at_idx" to table: "auth_action_tokens"
CREATE INDEX "auth_action_tokens_deleted_at_idx" ON "auth_action_tokens" ("deleted_at");
-- Create index "auth_action_tokens_expires_at_idx" to table: "auth_action_tokens"
CREATE INDEX "auth_action_tokens_expires_at_idx" ON "auth_action_tokens" ("expires_at") WHERE (deleted_at IS NULL);
-- Create index "auth_action_tokens_purpose_idx" to table: "auth_action_tokens"
CREATE INDEX "auth_action_tokens_purpose_idx" ON "auth_action_tokens" ("purpose") WHERE (deleted_at IS NULL);
-- Create index "auth_action_tokens_user_id_idx" to table: "auth_action_tokens"
CREATE INDEX "auth_action_tokens_user_id_idx" ON "auth_action_tokens" ("user_id") WHERE (deleted_at IS NULL);
-- Create trigger "auth_action_tokens_set_updated_at"
CREATE TRIGGER "auth_action_tokens_set_updated_at" BEFORE UPDATE ON "auth_action_tokens" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
