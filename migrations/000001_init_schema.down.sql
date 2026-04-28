-- 000001_init_schema.down.sql

BEGIN;

DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;

-- Keep extension (harmless and may be used by other schemas).
-- DROP EXTENSION IF EXISTS pgcrypto;

COMMIT;

