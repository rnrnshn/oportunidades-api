-- Modify "universities" table
ALTER TABLE "universities" ADD CONSTRAINT "universities_founded_year_check" CHECK ((founded_year IS NULL) OR (founded_year > 0)), ADD CONSTRAINT "universities_student_count_check" CHECK ((student_count IS NULL) OR (student_count >= 0)), ADD COLUMN "city" text NULL, ADD COLUMN "country" text NOT NULL DEFAULT 'Mozambique', ADD COLUMN "campus_image_url" text NULL, ADD COLUMN "founded_year" integer NULL, ADD COLUMN "address" text NULL, ADD COLUMN "map_url" text NULL, ADD COLUMN "academic_calendar" text NULL, ADD COLUMN "student_count" integer NULL, ADD COLUMN "admissions_deadline" date NULL, ADD COLUMN "tags" text[] NOT NULL DEFAULT '{}';
-- Create "university_fees" table
CREATE TABLE "university_fees" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "university_id" uuid NOT NULL,
  "label" text NOT NULL,
  "value" text NOT NULL,
  "sort_order" integer NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "university_fees_university_id_fkey" FOREIGN KEY ("university_id") REFERENCES "universities" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "university_fees_deleted_at_idx" to table: "university_fees"
CREATE INDEX "university_fees_deleted_at_idx" ON "university_fees" ("deleted_at");
-- Create index "university_fees_sort_order_idx" to table: "university_fees"
CREATE INDEX "university_fees_sort_order_idx" ON "university_fees" ("sort_order") WHERE (deleted_at IS NULL);
-- Create index "university_fees_university_id_idx" to table: "university_fees"
CREATE INDEX "university_fees_university_id_idx" ON "university_fees" ("university_id") WHERE (deleted_at IS NULL);
-- Create trigger "university_fees_set_updated_at"
CREATE TRIGGER "university_fees_set_updated_at" BEFORE UPDATE ON "university_fees" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
-- Create "university_scholarships" table
CREATE TABLE "university_scholarships" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "university_id" uuid NOT NULL,
  "name" text NOT NULL,
  "amount" text NULL,
  "status" text NOT NULL,
  "sort_order" integer NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "university_scholarships_university_id_fkey" FOREIGN KEY ("university_id") REFERENCES "universities" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "university_scholarships_deleted_at_idx" to table: "university_scholarships"
CREATE INDEX "university_scholarships_deleted_at_idx" ON "university_scholarships" ("deleted_at");
-- Create index "university_scholarships_sort_order_idx" to table: "university_scholarships"
CREATE INDEX "university_scholarships_sort_order_idx" ON "university_scholarships" ("sort_order") WHERE (deleted_at IS NULL);
-- Create index "university_scholarships_university_id_idx" to table: "university_scholarships"
CREATE INDEX "university_scholarships_university_id_idx" ON "university_scholarships" ("university_id") WHERE (deleted_at IS NULL);
-- Create trigger "university_scholarships_set_updated_at"
CREATE TRIGGER "university_scholarships_set_updated_at" BEFORE UPDATE ON "university_scholarships" FOR EACH ROW EXECUTE FUNCTION "set_updated_at"();
