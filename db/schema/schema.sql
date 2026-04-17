CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'mentor', 'cms_partner', 'admin')),
  name TEXT NOT NULL,
  avatar_url TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE INDEX users_role_idx ON users (role) WHERE deleted_at IS NULL;
CREATE INDEX users_deleted_at_idx ON users (deleted_at);

CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  revoked_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ,
  CONSTRAINT refresh_tokens_token_hash_unique UNIQUE (token_hash)
);

CREATE INDEX refresh_tokens_user_id_idx ON refresh_tokens (user_id) WHERE deleted_at IS NULL;
CREATE INDEX refresh_tokens_expires_at_idx ON refresh_tokens (expires_at) WHERE deleted_at IS NULL;
CREATE INDEX refresh_tokens_deleted_at_idx ON refresh_tokens (deleted_at);

CREATE TABLE universities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT NOT NULL,
  name TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('publica', 'privada', 'instituto', 'academia')),
  province TEXT NOT NULL,
  description TEXT,
  logo_url TEXT,
  website TEXT,
  email TEXT,
  phone TEXT,
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  verified_at TIMESTAMPTZ,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  CONSTRAINT universities_slug_unique UNIQUE (slug)
);

CREATE INDEX universities_name_idx ON universities (name) WHERE deleted_at IS NULL;
CREATE INDEX universities_province_idx ON universities (province) WHERE deleted_at IS NULL;
CREATE INDEX universities_type_idx ON universities (type) WHERE deleted_at IS NULL;
CREATE INDEX universities_verified_idx ON universities (verified) WHERE deleted_at IS NULL;
CREATE INDEX universities_deleted_at_idx ON universities (deleted_at);

CREATE TABLE courses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT NOT NULL,
  university_id UUID NOT NULL REFERENCES universities(id),
  name TEXT NOT NULL,
  area TEXT NOT NULL,
  level TEXT NOT NULL CHECK (level IN ('licenciatura', 'mestrado', 'doutoramento', 'tecnico_medio', 'cet')),
  regime TEXT NOT NULL CHECK (regime IN ('presencial', 'distancia', 'misto')),
  duration_years INTEGER CHECK (duration_years IS NULL OR duration_years > 0),
  annual_fee NUMERIC(12,2) CHECK (annual_fee IS NULL OR annual_fee >= 0),
  entry_requirements TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  CONSTRAINT courses_slug_unique UNIQUE (slug)
);

CREATE INDEX courses_university_id_idx ON courses (university_id) WHERE deleted_at IS NULL;
CREATE INDEX courses_area_idx ON courses (area) WHERE deleted_at IS NULL;
CREATE INDEX courses_level_idx ON courses (level) WHERE deleted_at IS NULL;
CREATE INDEX courses_regime_idx ON courses (regime) WHERE deleted_at IS NULL;
CREATE INDEX courses_deleted_at_idx ON courses (deleted_at);

CREATE TABLE opportunities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT NOT NULL,
  title TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('bolsa', 'estagio', 'emprego', 'intercambio', 'workshop', 'competicao')),
  entity_name TEXT NOT NULL,
  description TEXT NOT NULL,
  requirements TEXT,
  deadline TIMESTAMPTZ,
  apply_url TEXT,
  country TEXT NOT NULL,
  language TEXT,
  area TEXT,
  is_active BOOLEAN NOT NULL DEFAULT FALSE,
  published_by UUID REFERENCES users(id),
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  CONSTRAINT opportunities_slug_unique UNIQUE (slug)
);

CREATE INDEX opportunities_type_idx ON opportunities (type) WHERE deleted_at IS NULL;
CREATE INDEX opportunities_area_idx ON opportunities (area) WHERE deleted_at IS NULL;
CREATE INDEX opportunities_country_idx ON opportunities (country) WHERE deleted_at IS NULL;
CREATE INDEX opportunities_deadline_idx ON opportunities (deadline) WHERE deleted_at IS NULL;
CREATE INDEX opportunities_is_active_idx ON opportunities (is_active) WHERE deleted_at IS NULL;
CREATE INDEX opportunities_verified_idx ON opportunities (verified) WHERE deleted_at IS NULL;
CREATE INDEX opportunities_deleted_at_idx ON opportunities (deleted_at);

CREATE TABLE reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  reporter_id UUID REFERENCES users(id),
  entity_type TEXT NOT NULL CHECK (entity_type IN ('university', 'course', 'opportunity')),
  entity_id UUID NOT NULL,
  reason TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved', 'dismissed')),
  resolved_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX reports_reporter_id_idx ON reports (reporter_id) WHERE deleted_at IS NULL;
CREATE INDEX reports_entity_lookup_idx ON reports (entity_type, entity_id) WHERE deleted_at IS NULL;
CREATE INDEX reports_status_idx ON reports (status) WHERE deleted_at IS NULL;
CREATE INDEX reports_deleted_at_idx ON reports (deleted_at);

CREATE TRIGGER users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER refresh_tokens_set_updated_at
BEFORE UPDATE ON refresh_tokens
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER universities_set_updated_at
BEFORE UPDATE ON universities
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER courses_set_updated_at
BEFORE UPDATE ON courses
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER opportunities_set_updated_at
BEFORE UPDATE ON opportunities
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER reports_set_updated_at
BEFORE UPDATE ON reports
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
