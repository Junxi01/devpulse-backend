-- 000002_users_auth_fields.down.sql

BEGIN;

-- Re-add legacy column if needed for rollback.
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS avatar_url text NOT NULL DEFAULT '';

ALTER TABLE users
  DROP COLUMN IF EXISTS password_hash;

COMMIT;

