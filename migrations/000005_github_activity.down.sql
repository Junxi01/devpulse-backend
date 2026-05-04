-- 000005_github_activity.down.sql

BEGIN;

DROP TABLE IF EXISTS commits;
DROP TABLE IF EXISTS issues;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS repository_events;

COMMIT;
