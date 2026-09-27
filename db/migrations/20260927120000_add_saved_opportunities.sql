-- Create "saved_opportunities" table
CREATE TABLE "saved_opportunities" (
  "user_id" uuid NOT NULL,
  "opportunity_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("user_id", "opportunity_id"),
  CONSTRAINT "saved_opportunities_opportunity_id_fkey" FOREIGN KEY ("opportunity_id") REFERENCES "opportunities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "saved_opportunities_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "saved_opportunities_user_created_idx" to table: "saved_opportunities"
CREATE INDEX "saved_opportunities_user_created_idx" ON "saved_opportunities" ("user_id", "created_at" DESC);
