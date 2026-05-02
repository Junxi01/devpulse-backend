-- 000004_projects_repositories.up.sql
-- Workspace-scoped projects and repositories (Day 9). Replaces legacy user-owned projects.

BEGIN;

DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;

CREATE TABLE IF NOT EXISTS projects (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

  name         text NOT NULL,
  description  text NOT NULL DEFAULT '',

  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT projects_workspace_name_unique UNIQUE (workspace_id, name)
);

CREATE INDEX IF NOT EXISTS projects_workspace_id_idx ON projects(workspace_id);

CREATE TABLE IF NOT EXISTS repositories (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id      uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

  provider        text NOT NULL,
  owner           text NOT NULL,
  name            text NOT NULL,
  full_name       text NOT NULL,
  external_id     text NOT NULL,
  default_branch  text NOT NULL DEFAULT 'main',

  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT repositories_project_provider_external_unique UNIQUE (project_id, provider, external_id)
);

CREATE INDEX IF NOT EXISTS repositories_project_id_idx ON repositories(project_id);

COMMIT;
