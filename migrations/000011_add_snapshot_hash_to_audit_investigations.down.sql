DROP INDEX IF EXISTS idx_audit_investigations_snapshot_hash;

ALTER TABLE audit_investigations
DROP COLUMN IF EXISTS snapshot_hash;
