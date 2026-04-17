-- Create "set_updated_at" function
CREATE FUNCTION "set_updated_at" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;
-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "role" text NOT NULL DEFAULT 'user',
  "name" text NOT NULL,
  "avatar_url" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_unique" UNIQUE ("email"),
  CONSTRAINT "users_role_check" CHECK (role = ANY (ARRAY['user'::text, 'mentor'::text, 'cms_partner'::text, 'admin'::text]))
);
-- Create index "users_deleted_at_idx" to table: "users"
CREATE INDEX "users_deleted_at_idx" ON "users" ("deleted_at");
-- Create index "users_role_idx" to table: "users"
CREATE INDEX "users_role_idx" ON "users" ("role") WHERE (deleted_at IS NULL);
-- Create "universities" table
CREATE TABLE "universities" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "slug" text NOT NULL,
  "name" text NOT NULL,
  "type" text NOT NULL,
  "province" text NOT NULL,
  "description" text NULL,
  "logo_url" text NULL,
  "website" text NULL,
  "email" text NULL,
  "phone" text NULL,
  "verified" boolean NOT NULL DEFAULT false,
  "verified_at" timestamptz NULL,
  "created_by" uuid NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "universities_slug_unique" UNIQUE ("slug"),
  CONSTRAINT "universities_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "universities_type_check" CHECK (type = ANY (ARRAY['publica'::text, 'privada'::text, 'instituto'::text, 'academia'::text]))
);
-- Create index "universities_deleted_at_idx" to table: "universities"
CREATE INDEX "universities_deleted_at_idx" ON "universities" ("deleted_at");
-- Create index "universities_name_idx" to table: "universities"
CREATE INDEX "universities_name_idx" ON "universities" ("name") WHERE (deleted_at IS NULL);
-- Create index "universities_province_idx" to table: "universities"
CREATE INDEX "universities_province_idx" ON "universities" ("province") WHERE (deleted_at IS NULL);
-- Create index "universities_type_idx" to table: "universities"
CREATE INDEX "universities_type_idx" ON "universities" ("type") WHERE (deleted_at IS NULL);
-- Create index "universities_verified_idx" to table: "universities"
CREATE INDEX "universities_verified_idx" ON "universities" ("verified") WHERE (deleted_at IS NULL);
-- Create trigger "universities_set_updated_at"
CREATE TRIGGER "universities_set_updated_at" BEFORE UPDATE ON "universities" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create "courses" table
CREATE TABLE "courses" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "slug" text NOT NULL,
  "university_id" uuid NOT NULL,
  "name" text NOT NULL,
  "area" text NOT NULL,
  "level" text NOT NULL,
  "regime" text NOT NULL,
  "duration_years" integer NULL,
  "annual_fee" numeric(12,2) NULL,
  "entry_requirements" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "courses_slug_unique" UNIQUE ("slug"),
  CONSTRAINT "courses_university_id_fkey" FOREIGN KEY ("university_id") REFERENCES "universities" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "courses_annual_fee_check" CHECK ((annual_fee IS NULL) OR (annual_fee >= (0)::numeric)),
  CONSTRAINT "courses_duration_years_check" CHECK ((duration_years IS NULL) OR (duration_years > 0)),
  CONSTRAINT "courses_level_check" CHECK (level = ANY (ARRAY['licenciatura'::text, 'mestrado'::text, 'doutoramento'::text, 'tecnico_medio'::text, 'cet'::text])),
  CONSTRAINT "courses_regime_check" CHECK (regime = ANY (ARRAY['presencial'::text, 'distancia'::text, 'misto'::text]))
);
-- Create index "courses_area_idx" to table: "courses"
CREATE INDEX "courses_area_idx" ON "courses" ("area") WHERE (deleted_at IS NULL);
-- Create index "courses_deleted_at_idx" to table: "courses"
CREATE INDEX "courses_deleted_at_idx" ON "courses" ("deleted_at");
-- Create index "courses_level_idx" to table: "courses"
CREATE INDEX "courses_level_idx" ON "courses" ("level") WHERE (deleted_at IS NULL);
-- Create index "courses_regime_idx" to table: "courses"
CREATE INDEX "courses_regime_idx" ON "courses" ("regime") WHERE (deleted_at IS NULL);
-- Create index "courses_university_id_idx" to table: "courses"
CREATE INDEX "courses_university_id_idx" ON "courses" ("university_id") WHERE (deleted_at IS NULL);
-- Create trigger "courses_set_updated_at"
CREATE TRIGGER "courses_set_updated_at" BEFORE UPDATE ON "courses" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create "opportunities" table
CREATE TABLE "opportunities" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "slug" text NOT NULL,
  "title" text NOT NULL,
  "type" text NOT NULL,
  "entity_name" text NOT NULL,
  "description" text NOT NULL,
  "requirements" text NULL,
  "deadline" timestamptz NULL,
  "apply_url" text NULL,
  "country" text NOT NULL,
  "language" text NULL,
  "area" text NULL,
  "is_active" boolean NOT NULL DEFAULT false,
  "published_by" uuid NULL,
  "verified" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "opportunities_slug_unique" UNIQUE ("slug"),
  CONSTRAINT "opportunities_published_by_fkey" FOREIGN KEY ("published_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "opportunities_type_check" CHECK (type = ANY (ARRAY['bolsa'::text, 'estagio'::text, 'emprego'::text, 'intercambio'::text, 'workshop'::text, 'competicao'::text]))
);
-- Create index "opportunities_area_idx" to table: "opportunities"
CREATE INDEX "opportunities_area_idx" ON "opportunities" ("area") WHERE (deleted_at IS NULL);
-- Create index "opportunities_country_idx" to table: "opportunities"
CREATE INDEX "opportunities_country_idx" ON "opportunities" ("country") WHERE (deleted_at IS NULL);
-- Create index "opportunities_deadline_idx" to table: "opportunities"
CREATE INDEX "opportunities_deadline_idx" ON "opportunities" ("deadline") WHERE (deleted_at IS NULL);
-- Create index "opportunities_deleted_at_idx" to table: "opportunities"
CREATE INDEX "opportunities_deleted_at_idx" ON "opportunities" ("deleted_at");
-- Create index "opportunities_is_active_idx" to table: "opportunities"
CREATE INDEX "opportunities_is_active_idx" ON "opportunities" ("is_active") WHERE (deleted_at IS NULL);
-- Create index "opportunities_type_idx" to table: "opportunities"
CREATE INDEX "opportunities_type_idx" ON "opportunities" ("type") WHERE (deleted_at IS NULL);
-- Create index "opportunities_verified_idx" to table: "opportunities"
CREATE INDEX "opportunities_verified_idx" ON "opportunities" ("verified") WHERE (deleted_at IS NULL);
-- Create trigger "opportunities_set_updated_at"
CREATE TRIGGER "opportunities_set_updated_at" BEFORE UPDATE ON "opportunities" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create "refresh_tokens" table
CREATE TABLE "refresh_tokens" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "token_hash" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "revoked_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "refresh_tokens_token_hash_unique" UNIQUE ("token_hash"),
  CONSTRAINT "refresh_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "refresh_tokens_deleted_at_idx" to table: "refresh_tokens"
CREATE INDEX "refresh_tokens_deleted_at_idx" ON "refresh_tokens" ("deleted_at");
-- Create index "refresh_tokens_expires_at_idx" to table: "refresh_tokens"
CREATE INDEX "refresh_tokens_expires_at_idx" ON "refresh_tokens" ("expires_at") WHERE (deleted_at IS NULL);
-- Create index "refresh_tokens_user_id_idx" to table: "refresh_tokens"
CREATE INDEX "refresh_tokens_user_id_idx" ON "refresh_tokens" ("user_id") WHERE (deleted_at IS NULL);
-- Create trigger "refresh_tokens_set_updated_at"
CREATE TRIGGER "refresh_tokens_set_updated_at" BEFORE UPDATE ON "refresh_tokens" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create "reports" table
CREATE TABLE "reports" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "reporter_id" uuid NULL,
  "entity_type" text NOT NULL,
  "entity_id" uuid NOT NULL,
  "reason" text NOT NULL,
  "status" text NOT NULL DEFAULT 'pending',
  "resolved_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "reports_reporter_id_fkey" FOREIGN KEY ("reporter_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "reports_entity_type_check" CHECK (entity_type = ANY (ARRAY['university'::text, 'course'::text, 'opportunity'::text])),
  CONSTRAINT "reports_status_check" CHECK (status = ANY (ARRAY['pending'::text, 'reviewed'::text, 'resolved'::text, 'dismissed'::text]))
);
-- Create index "reports_deleted_at_idx" to table: "reports"
CREATE INDEX "reports_deleted_at_idx" ON "reports" ("deleted_at");
-- Create index "reports_entity_lookup_idx" to table: "reports"
CREATE INDEX "reports_entity_lookup_idx" ON "reports" ("entity_type", "entity_id") WHERE (deleted_at IS NULL);
-- Create index "reports_reporter_id_idx" to table: "reports"
CREATE INDEX "reports_reporter_id_idx" ON "reports" ("reporter_id") WHERE (deleted_at IS NULL);
-- Create index "reports_status_idx" to table: "reports"
CREATE INDEX "reports_status_idx" ON "reports" ("status") WHERE (deleted_at IS NULL);
-- Create trigger "reports_set_updated_at"
CREATE TRIGGER "reports_set_updated_at" BEFORE UPDATE ON "reports" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create trigger "users_set_updated_at"
CREATE TRIGGER "users_set_updated_at" BEFORE UPDATE ON "users" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
