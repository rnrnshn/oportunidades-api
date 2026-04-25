-- Repeatable local seed data for development.
-- Login credentials for seeded users use password: password

INSERT INTO users (id, email, password_hash, role, name, avatar_url)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'admin@oportunidades.co.mz', '$2a$10$Oax5cYad3QkN1l6vn0pJ4O5ykwbYArlZdAOeC9J5RNbm2xw3bDwvG', 'admin', 'Admin Oportunidades', NULL),
  ('22222222-2222-2222-2222-222222222222', 'user@oportunidades.co.mz', '$2a$10$Oax5cYad3QkN1l6vn0pJ4O5ykwbYArlZdAOeC9J5RNbm2xw3bDwvG', 'user', 'Utilizador Demo', NULL),
  ('99999999-9999-9999-9999-999999999999', 'mentor@oportunidades.co.mz', '$2a$10$Oax5cYad3QkN1l6vn0pJ4O5ykwbYArlZdAOeC9J5RNbm2xw3bDwvG', 'mentor', 'Mentor Demo', NULL)
ON CONFLICT (email) DO UPDATE
SET
  password_hash = EXCLUDED.password_hash,
  role = EXCLUDED.role,
  name = EXCLUDED.name,
  avatar_url = EXCLUDED.avatar_url,
  deleted_at = NULL;

INSERT INTO universities (id, slug, name, type, province, city, country, description, website, email, phone, founded_year, address, academic_calendar, student_count, admissions_deadline, tags, verified, verified_at, created_by)
VALUES
  ('33333333-3333-3333-3333-333333333333', 'universidade-eduardo-mondlane', 'Universidade Eduardo Mondlane', 'publica', 'Maputo', 'Maputo', 'Mozambique', 'Universidade publica de referencia em Mocambique, reconhecida pela investigacao em ciencias naturais, saude publica e economia.', 'https://www.uem.mz', 'info@uem.mz', '+25821000001', 1962, 'Av. Julius Nyerere, Maputo', 'Janeiro - Novembro', 25000, '2026-09-15', ARRAY['Government', 'Academics', 'Tuition'], TRUE, NOW(), '11111111-1111-1111-1111-111111111111'),
  ('44444444-4444-4444-4444-444444444444', 'instituto-superior-de-transportes', 'Instituto Superior de Transportes e Comunicacoes', 'instituto', 'Maputo', 'Maputo', 'Mozambique', 'Instituicao focada em engenharia, tecnologia e logistica.', 'https://www.isutc.ac.mz', 'geral@isutc.ac.mz', '+25821000002', 1999, 'Av. Kim Il Sung, Maputo', 'Fevereiro - Dezembro', 3500, '2026-08-30', ARRAY['Engineering', 'Logistics', 'Technology'], TRUE, NOW(), '11111111-1111-1111-1111-111111111111')
ON CONFLICT (slug) DO UPDATE
SET
  name = EXCLUDED.name,
  type = EXCLUDED.type,
  province = EXCLUDED.province,
  city = EXCLUDED.city,
  country = EXCLUDED.country,
  description = EXCLUDED.description,
  website = EXCLUDED.website,
  email = EXCLUDED.email,
  phone = EXCLUDED.phone,
  founded_year = EXCLUDED.founded_year,
  address = EXCLUDED.address,
  academic_calendar = EXCLUDED.academic_calendar,
  student_count = EXCLUDED.student_count,
  admissions_deadline = EXCLUDED.admissions_deadline,
  tags = EXCLUDED.tags,
  verified = EXCLUDED.verified,
  verified_at = EXCLUDED.verified_at,
  created_by = EXCLUDED.created_by,
  deleted_at = NULL;

INSERT INTO university_fees (id, university_id, label, value, sort_order)
VALUES
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1', '33333333-3333-3333-3333-333333333333', 'Propina anual', '45,000 MZN', 0),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2', '33333333-3333-3333-3333-333333333333', 'Taxa de inscricao', '1,500 MZN', 1),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb3', '33333333-3333-3333-3333-333333333333', 'Prazos', '15 Setembro 2026', 2),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb4', '44444444-4444-4444-4444-444444444444', 'Propina anual', '52,000 MZN', 0),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb5', '44444444-4444-4444-4444-444444444444', 'Taxa de inscricao', '2,000 MZN', 1)
ON CONFLICT (id) DO UPDATE
SET
  university_id = EXCLUDED.university_id,
  label = EXCLUDED.label,
  value = EXCLUDED.value,
  sort_order = EXCLUDED.sort_order,
  deleted_at = NULL;

INSERT INTO university_scholarships (id, university_id, name, amount, status, sort_order)
VALUES
  ('cccccccc-cccc-cccc-cccc-ccccccccccc1', '33333333-3333-3333-3333-333333333333', 'STEM Excellence', '25,000 MZN', 'Aberta', 0),
  ('cccccccc-cccc-cccc-cccc-ccccccccccc2', '33333333-3333-3333-3333-333333333333', 'Eduardo Mondlane Alumni', '20,000 MZN', 'Inscricoes em breve', 1),
  ('cccccccc-cccc-cccc-cccc-ccccccccccc3', '44444444-4444-4444-4444-444444444444', 'Bolsa Engenharia Aplicada', '15,000 MZN', 'Aberta', 0)
ON CONFLICT (id) DO UPDATE
SET
  university_id = EXCLUDED.university_id,
  name = EXCLUDED.name,
  amount = EXCLUDED.amount,
  status = EXCLUDED.status,
  sort_order = EXCLUDED.sort_order,
  deleted_at = NULL;

INSERT INTO mentor_profiles (id, user_id, headline, bio, expertise, availability, is_active)
VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '99999999-9999-9999-9999-999999999999', 'Engenheiro de Software e mentor de carreira', 'Profissional com experiencia em desenvolvimento backend, transicao de carreira e preparacao para entrevistas.', 'Go, SQL, APIs, Carreira em Tecnologia', 'Sabados de manha e quartas a noite', TRUE)
ON CONFLICT (user_id) DO UPDATE
SET
  headline = EXCLUDED.headline,
  bio = EXCLUDED.bio,
  expertise = EXCLUDED.expertise,
  availability = EXCLUDED.availability,
  is_active = EXCLUDED.is_active,
  deleted_at = NULL;

INSERT INTO courses (id, slug, university_id, name, area, level, regime, duration_years, annual_fee, entry_requirements)
VALUES
  ('55555555-5555-5555-5555-555555555555', 'engenharia-informatica-uem', '33333333-3333-3333-3333-333333333333', 'Engenharia Informatica', 'Tecnologia', 'licenciatura', 'presencial', 4, 45000.00, '12a classe com Matematica e Fisica.'),
  ('66666666-6666-6666-6666-666666666666', 'logistica-e-transportes-isutc', '44444444-4444-4444-4444-444444444444', 'Logistica e Transportes', 'Gestao', 'licenciatura', 'presencial', 4, 52000.00, '12a classe concluida.')
ON CONFLICT (slug) DO UPDATE
SET
  university_id = EXCLUDED.university_id,
  name = EXCLUDED.name,
  area = EXCLUDED.area,
  level = EXCLUDED.level,
  regime = EXCLUDED.regime,
  duration_years = EXCLUDED.duration_years,
  annual_fee = EXCLUDED.annual_fee,
  entry_requirements = EXCLUDED.entry_requirements,
  deleted_at = NULL;

INSERT INTO opportunities (id, slug, title, type, entity_name, description, requirements, deadline, apply_url, external_url_label, country, location, is_remote, language, area, amount_min, amount_max, amount_currency, coverage, eligibility, application_process, degree_level, program_area, is_active, published_by, verified)
VALUES
  ('77777777-7777-7777-7777-777777777777', 'bolsa-mestrado-data-science', 'Bolsa de Mestrado em Data Science', 'bolsa', 'Fundacao Oportunidades', 'Bolsa integral para estudantes com interesse em ciencia de dados e impacto social.', 'Licenciatura em area relevante, ingles funcional e forte desempenho academico.', NOW() + INTERVAL '45 days', 'https://oportunidades.co.mz/bolsa-data-science', 'Visit scholarship website', 'Mozambique', 'Maputo', FALSE, 'Portugues', 'Tecnologia', 25000.00, 100000.00, 'MZN', ARRAY['Propina', 'Alojamento', 'Bolsa mensal'], 'Licenciatura concluida em area STEM. Media minima de 14 valores. Demonstrar necessidade financeira e motivacao para ciencia de dados.', '1. Submeter formulario online. 2. Anexar certificado e historico academico. 3. Enviar carta de motivacao. 4. Participar em entrevista final.', 'Mestrado', 'Data Science', TRUE, '11111111-1111-1111-1111-111111111111', TRUE),
  ('88888888-8888-8888-8888-888888888888', 'estagio-software-maputo', 'Estagio em Desenvolvimento de Software', 'estagio', 'Tech MZ', 'Programa de estagio para recem-graduados em engenharia informatica.', 'Conhecimentos de Go, SQL e Git.', NOW() + INTERVAL '20 days', 'https://oportunidades.co.mz/estagio-software', NULL, 'Mozambique', 'Maputo', FALSE, 'Portugues', 'Tecnologia', NULL, NULL, 'MZN', ARRAY[]::text[], NULL, NULL, NULL, NULL, TRUE, '11111111-1111-1111-1111-111111111111', TRUE)
ON CONFLICT (slug) DO UPDATE
SET
  title = EXCLUDED.title,
  type = EXCLUDED.type,
  entity_name = EXCLUDED.entity_name,
  description = EXCLUDED.description,
  requirements = EXCLUDED.requirements,
  deadline = EXCLUDED.deadline,
  apply_url = EXCLUDED.apply_url,
  external_url_label = EXCLUDED.external_url_label,
  country = EXCLUDED.country,
  location = EXCLUDED.location,
  is_remote = EXCLUDED.is_remote,
  language = EXCLUDED.language,
  area = EXCLUDED.area,
  amount_min = EXCLUDED.amount_min,
  amount_max = EXCLUDED.amount_max,
  amount_currency = EXCLUDED.amount_currency,
  coverage = EXCLUDED.coverage,
  eligibility = EXCLUDED.eligibility,
  application_process = EXCLUDED.application_process,
  degree_level = EXCLUDED.degree_level,
  program_area = EXCLUDED.program_area,
  is_active = EXCLUDED.is_active,
  published_by = EXCLUDED.published_by,
  verified = EXCLUDED.verified,
  deleted_at = NULL;

INSERT INTO articles (id, slug, title, excerpt, content, cover_image_url, type, status, source_name, source_url, seo_title, seo_description, is_featured, author_id, published_at)
VALUES
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'como-preparar-candidatura-a-bolsa', 'Como Preparar uma Candidatura a Bolsa', 'Guia pratico para estudantes que querem submeter candidaturas mais fortes.', 'Este guia resume como organizar documentos, escrever motivacao e rever requisitos antes do prazo final.', NULL, 'guide', 'published', 'Oportunidades', 'https://oportunidades.co.mz/guias/candidatura-bolsa', 'Como preparar candidatura a bolsa', 'Guia pratico para melhorar a sua candidatura a bolsas.', TRUE, '11111111-1111-1111-1111-111111111111', NOW() - INTERVAL '2 days'),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'novas-bolsas-lancadas-para-2026', 'Novas Bolsas Lancadas para 2026', 'Resumo das novas oportunidades de bolsa abertas este mes.', 'Compilamos as principais bolsas anunciadas para 2026 com foco em tecnologia, saude e gestao.', NULL, 'news', 'published', 'Oportunidades', 'https://oportunidades.co.mz/noticias/bolsas-2026', 'Novas bolsas lancadas para 2026', 'Resumo das bolsas recentemente anunciadas para 2026.', FALSE, '11111111-1111-1111-1111-111111111111', NOW() - INTERVAL '1 day'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'porque-mentoria-importa-na-carreira', 'Porque a Mentoria Importa na Carreira', 'Uma reflexao editorial sobre o valor da mentoria para jovens profissionais.', 'Mentoria acelera aprendizagem, expande contexto e ajuda jovens profissionais a evitar erros comuns no inicio da carreira.', NULL, 'editorial', 'published', 'Oportunidades', 'https://oportunidades.co.mz/editorial/mentoria-carreira', 'Porque a mentoria importa na carreira', 'Editorial sobre o impacto da mentoria no crescimento profissional.', TRUE, '11111111-1111-1111-1111-111111111111', NOW() - INTERVAL '3 days')
ON CONFLICT (slug) DO UPDATE
SET
  title = EXCLUDED.title,
  excerpt = EXCLUDED.excerpt,
  content = EXCLUDED.content,
  cover_image_url = EXCLUDED.cover_image_url,
  type = EXCLUDED.type,
  status = EXCLUDED.status,
  source_name = EXCLUDED.source_name,
  source_url = EXCLUDED.source_url,
  seo_title = EXCLUDED.seo_title,
  seo_description = EXCLUDED.seo_description,
  is_featured = EXCLUDED.is_featured,
  author_id = EXCLUDED.author_id,
  published_at = EXCLUDED.published_at,
  deleted_at = NULL;
