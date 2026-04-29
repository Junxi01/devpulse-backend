-- 000002_users_auth_fields.up.sql
-- Ensure users table has auth-related fields required by Day 4.

BEGIN;

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS password_hash text NOT NULL DEFAULT '';

-- Align Day 4 users fields (optional older column).
ALTER TABLE users
  DROP COLUMN IF EXISTS avatar_url;

COMMIT;

