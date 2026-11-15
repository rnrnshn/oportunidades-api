-- Modify "users" table: account suspension
ALTER TABLE "users" ADD COLUMN "suspended_at" timestamptz NULL, ADD COLUMN "suspended_by" uuid NULL, ADD COLUMN "suspension_reason" text NULL, ADD CONSTRAINT "users_suspended_by_fkey" FOREIGN KEY ("suspended_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create "user_invites" table
CREATE TABLE "user_invites" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" text NOT NULL,
  "name" text NOT NULL,
  "role" text NOT NULL,
  "token_hash" text NOT NULL,
  "invited_by" uuid NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "accepted_at" timestamptz NULL,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "user_invites_token_hash_unique" UNIQUE ("token_hash"),
  CONSTRAINT "user_invites_invited_by_fkey" FOREIGN KEY ("invited_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "user_invites_role_check" CHECK (role = ANY (ARRAY['admin'::text, 'cms_partner'::text]))
);
-- Create index "user_invites_pending_idx" to table: "user_invites"
CREATE INDEX "user_invites_pending_idx" ON "user_invites" ((lower(email))) WHERE ((accepted_at IS NULL) AND (revoked_at IS NULL));
-- Create "audit_events" table
CREATE TABLE "audit_events" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "actor_id" uuid NULL,
  "action" text NOT NULL,
  "target_user_id" uuid NULL,
  "data" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "audit_events_actor_id_fkey" FOREIGN KEY ("actor_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "audit_events_target_user_id_fkey" FOREIGN KEY ("target_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "audit_events_action_check" CHECK (action = ANY (ARRAY['role_changed'::text, 'suspended'::text, 'unsuspended'::text, 'invite_sent'::text, 'invite_resent'::text, 'invite_revoked'::text, 'invite_accepted'::text]))
);
-- Create index "audit_events_target_idx" to table: "audit_events"
CREATE INDEX "audit_events_target_idx" ON "audit_events" ("target_user_id", "created_at" DESC);
