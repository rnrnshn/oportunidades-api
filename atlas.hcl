env "dev" {
  url = getenv("DATABASE_URL")
  src = "file://db/schema/schema.sql"
  dev = "docker://postgres/16/dev?search_path=public"
  migration {
    dir = "file://db/migrations"
  }
}
