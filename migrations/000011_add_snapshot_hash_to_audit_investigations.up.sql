ALTER TABLE audit_investigations
ADD COLUMN IF NOT EXISTS snapshot_hash TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_audit_investigations_snapshot_hash
ON audit_investigations (created_by_user_id, snapshot_hash, created_at DESC);
