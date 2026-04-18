# Oportunidades API

> API backend do ecossistema Oportunidades.

Esta API serve o frontend público, o admin UI e futuros clientes internos/mobile. O backend é construído em Go + Fiber, usa PostgreSQL com `sqlc`, migrações com Atlas e autenticação própria com JWT + refresh token rotation.

---

## Stack

```txt
Go 1.23+
Fiber v2
PostgreSQL
pgx/v5
sqlc
Atlas
zerolog
JWT + bcrypt
```

---

## Estado actual

Implementado:

- auth
  - register
  - login
  - refresh
  - logout
  - logout-all
  - forgot password
  - reset password
  - send verification
  - verify email
  - deactivate account
- account
  - `GET /v1/account/me`
  - `PATCH /v1/account/me`
  - `POST /v1/account/password`
- catalog público
  - universidades list/detail
  - cursos list/detail
- opportunities públicas
  - list/detail
- articles públicas
  - list/detail
- mentorship
  - mentors list/detail
  - session request
  - session list/detail/update
- reports
  - create
  - admin list/detail/update
- CMS
  - articles list/detail/create/edit
  - opportunities list/detail/create/edit
  - universities list/detail/create/edit
  - courses list/detail/create/edit
- admin lifecycle
  - publish/unpublish/archive article
  - verify/reject/deactivate opportunity
- uploads
  - presign
  - confirm
- local seed data
- centralized validation for write paths and list query params

---

## Estrutura

```txt
cmd/api/
internal/
  account/
  admin/
  articles/
  auth/
  catalog/
  cms/
  mentorship/
  opportunities/
  reports/
  uploads/
pkg/
  apierror/
  db/
  middleware/
  storage/
  validation/
db/
  schema/
  queries/
  migrations/
  seeds/
scripts/
```

---

## Requisitos

- Go 1.23+
- Docker
- pnpm not required for this repo

---

## Quick Start

```bash
docker compose up -d
cp .env.example .env

set -a
source ./.env
set +a

go run ./cmd/api
```

API local:

```txt
http://localhost:8080
```

Health check:

```bash
curl http://localhost:8080/health
```

---

## Seed Data

Aplicar seed local:

```bash
./scripts/seed-dev.sh
```

Credenciais seeded:

- admin
  - email: `admin@oportunidades.co.mz`
  - password: `password`
- mentor
  - email: `mentor@oportunidades.co.mz`
  - password: `password`
- user
  - email: `user@oportunidades.co.mz`
  - password: `password`

---

## Configuração

`.env.example`:

```env
PORT=8080
ENV=development
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/oportunidades?sslmode=disable
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

JWT_SECRET=change-me
JWT_EXPIRY_MINUTES=15
REFRESH_TOKEN_EXPIRY_DAYS=30
AUTH_ACTION_TOKEN_EXPIRY_HOURS=24

SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_ROLE_KEY=change-me
SUPABASE_STORAGE_BUCKET=oportunidades-media

RESEND_API_KEY=change-me
EMAIL_FROM=noreply@oportunidades.co.mz
```

Notas:

- `SUPABASE_*` é usado para presigned uploads/media
- `RESEND_*` está preparado para email delivery, mas os fluxos podem ainda usar `debug_token` em desenvolvimento

---

## Migrations

Gerar nova migration:

```bash
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/oportunidades?sslmode=disable"
~/.local/bin/atlas migrate diff nome_da_migration --env dev
```

Aplicar migrations:

```bash
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/oportunidades?sslmode=disable"
~/.local/bin/atlas migrate apply --env dev
```

Regenerar `sqlc`:

```bash
docker run --rm -e HOME=/tmp -v "$PWD:/src:Z" -w /src sqlc/sqlc:1.30.0 generate -f /src/sqlc.yaml
```

---

## Uploads

Fluxo actual:

1. `POST /v1/uploads/presign`
2. frontend faz upload directo para Supabase Storage
3. `POST /v1/uploads/confirm`

Pastas suportadas no bucket:

- `articles`
- `opportunities`
- `universities`
- `users`

---

## Testes

```bash
go test ./...
go test -race ./...
```

---

## Endpoints principais

### Auth

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout`
- `POST /v1/auth/forgot-password`
- `POST /v1/auth/reset-password`
- `POST /v1/auth/send-verification`
- `POST /v1/auth/verify-email`

### Account

- `GET /v1/account/me`
- `PATCH /v1/account/me`
- `POST /v1/account/password`
- `POST /v1/account/logout-all`
- `POST /v1/account/deactivate`

### Public

- `GET /v1/articles`
- `GET /v1/articles/:slug`
- `GET /v1/catalog/universities`
- `GET /v1/catalog/universities/:slug`
- `GET /v1/catalog/courses`
- `GET /v1/catalog/courses/:slug`
- `GET /v1/opportunities`
- `GET /v1/opportunities/:slug`
- `GET /v1/mentorship/mentors`
- `GET /v1/mentorship/mentors/:id`
- `POST /v1/mentorship/sessions`
- `GET /v1/mentorship/sessions`
- `GET /v1/mentorship/sessions/:id`
- `PATCH /v1/mentorship/sessions/:id`
- `POST /v1/reports`

### CMS

- `GET /v1/cms/articles`
- `GET /v1/cms/articles/:id`
- `POST /v1/cms/articles`
- `PATCH /v1/cms/articles/:id`
- `GET /v1/cms/opportunities`
- `GET /v1/cms/opportunities/:id`
- `POST /v1/cms/opportunities`
- `PATCH /v1/cms/opportunities/:id`
- `GET /v1/cms/universities`
- `GET /v1/cms/universities/:id`
- `POST /v1/cms/universities`
- `PATCH /v1/cms/universities/:id`
- `GET /v1/cms/courses`
- `GET /v1/cms/courses/:id`
- `POST /v1/cms/courses`
- `PATCH /v1/cms/courses/:id`

### Admin

- `POST /v1/admin/articles/:id/publish`
- `POST /v1/admin/articles/:id/unpublish`
- `POST /v1/admin/articles/:id/archive`
- `POST /v1/admin/opportunities/:id/verify`
- `POST /v1/admin/opportunities/:id/reject`
- `POST /v1/admin/opportunities/:id/deactivate`
- `GET /v1/admin/reports`
- `GET /v1/admin/reports/:id`
- `PATCH /v1/admin/reports/:id`

---

## Notas de implementação

- JWT access token + refresh token rotation
- CORS com `AllowCredentials` activo para admin/frontend local
- uploads usam Supabase Storage para o MVP
- articles suportam:
  - `content` HTML
  - `content_json` JSONB estruturado

---

## Próximos passos naturais

- email delivery real
- upload/media association polish
- report moderation notes/detail expansion
- public/frontend rendering de `content_json` quando for necessário
