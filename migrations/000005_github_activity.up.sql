-- 000005_github_activity.up.sql
-- GitHub-style activity for linked repositories (events, PRs, issues, commits).

BEGIN;

CREATE TABLE IF NOT EXISTS repository_events (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  repository_id  uuid NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,

  event_type     text NOT NULL,
  external_id    text NOT NULL,
  payload        jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at    timestamptz NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT repository_events_repo_external_unique UNIQUE (repository_id, external_id)
);

CREATE INDEX IF NOT EXISTS repository_events_repository_id_idx ON repository_events(repository_id);
CREATE INDEX IF NOT EXISTS repository_events_occurred_at_idx ON repository_events(occurred_at DESC);

CREATE TABLE IF NOT EXISTS pull_requests (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  repository_id   uuid NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,

  number          int NOT NULL,
  title           text NOT NULL,
  state           text NOT NULL,
  author          text NOT NULL,
  base_branch     text NOT NULL,
  head_branch     text NOT NULL,
  changed_files   int NOT NULL DEFAULT 0,
  additions       int NOT NULL DEFAULT 0,
  deletions       int NOT NULL DEFAULT 0,
  risk_level      text,

  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT pull_requests_repository_number_unique UNIQUE (repository_id, number)
);

CREATE INDEX IF NOT EXISTS pull_requests_repository_id_idx ON pull_requests(repository_id);

CREATE TABLE IF NOT EXISTS issues (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  repository_id   uuid NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,

  number          int NOT NULL,
  title           text NOT NULL,
  state           text NOT NULL,
  author          text NOT NULL,
  labels          jsonb NOT NULL DEFAULT '[]'::jsonb,
  priority        text,
  category        text,

  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT issues_repository_number_unique UNIQUE (repository_id, number)
);

CREATE INDEX IF NOT EXISTS issues_repository_id_idx ON issues(repository_id);

CREATE TABLE IF NOT EXISTS commits (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  repository_id   uuid NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,

  sha             text NOT NULL,
  message         text NOT NULL DEFAULT '',
  author          text NOT NULL,
  committed_at    timestamptz NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT commits_repository_sha_unique UNIQUE (repository_id, sha)
);

CREATE INDEX IF NOT EXISTS commits_repository_id_idx ON commits(repository_id);
CREATE INDEX IF NOT EXISTS commits_committed_at_idx ON commits(committed_at DESC);

COMMIT;
