CREATE TABLE applications (id bigint PRIMARY KEY, status text NOT NULL);
INSERT INTO applications VALUES (42, 'submitted');

CREATE TABLE audit_logs (
  id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  actor_id    bigint      NOT NULL,
  action      text        NOT NULL,
  target_type text        NOT NULL,
  target_id   bigint      NOT NULL,
  detail      jsonb       NOT NULL DEFAULT '{}'
);

CREATE FUNCTION forbid_audit_log_change() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION '監査ログは変更できません';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_logs_append_only
BEFORE UPDATE OR DELETE ON audit_logs
FOR EACH ROW EXECUTE FUNCTION forbid_audit_log_change();

BEGIN;
UPDATE applications SET status = 'approved' WHERE id = 42 AND status = 'submitted';
INSERT INTO audit_logs (actor_id, action, target_type, target_id, detail)
VALUES (7, 'application.approve', 'application', 42, '{"from": "submitted", "to": "approved"}');
COMMIT;

SELECT actor_id, action, target_type, target_id, detail FROM audit_logs;
DELETE FROM audit_logs WHERE id = 1;
UPDATE audit_logs SET actor_id = 8 WHERE id = 1;
