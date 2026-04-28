-- 000001_init_schema.up.sql
-- Core tables for DevPulse (Day 1): users, projects, project_members
-- UUID primary keys, timestamptz timestamps.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email        text NOT NULL UNIQUE,
  name         text NOT NULL DEFAULT '',
  avatar_url   text NOT NULL DEFAULT '',
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS projects (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id     uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

  -- Future-friendly identity fields (GitHub repo, webhooks, etc.)
  name         text NOT NULL,
  description  text NOT NULL DEFAULT '',

  github_owner text NOT NULL DEFAULT '',
  github_repo  text NOT NULL DEFAULT '',

  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT projects_owner_name_unique UNIQUE (owner_id, name)
);

CREATE INDEX IF NOT EXISTS projects_owner_id_idx ON projects(owner_id);
CREATE INDEX IF NOT EXISTS projects_github_repo_idx ON projects(github_owner, github_repo);

CREATE TABLE IF NOT EXISTS project_members (
  project_id   uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role         text NOT NULL DEFAULT 'member',
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),

  PRIMARY KEY (project_id, user_id)
);

CREATE INDEX IF NOT EXISTS project_members_user_id_idx ON project_members(user_id);

COMMIT;

