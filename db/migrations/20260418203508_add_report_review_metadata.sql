-- Modify "reports" table
ALTER TABLE "reports" ADD COLUMN "reviewed_by" uuid NULL, ADD COLUMN "moderation_notes" text NULL, ADD CONSTRAINT "reports_reviewed_by_fkey" FOREIGN KEY ("reviewed_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create index "reports_reviewed_by_idx" to table: "reports"
CREATE INDEX "reports_reviewed_by_idx" ON "reports" ("reviewed_by") WHERE (deleted_at IS NULL);
