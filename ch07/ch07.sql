\timing off
CREATE TABLE organizations (
  id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name text NOT NULL
);
CREATE TABLE applications (
  id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_id        bigint NOT NULL,
  organization_id bigint NOT NULL REFERENCES organizations (id),
  status          text   NOT NULL,
  created_at      timestamptz NOT NULL
);
INSERT INTO organizations (name) SELECT '団体' || i FROM generate_series(1, 2000) AS i;
INSERT INTO applications (event_id, organization_id, status, created_at)
SELECT (i % 50) + 1, (i % 2000) + 1,
       (ARRAY['draft', 'submitted', 'approved'])[(i % 3) + 1],
       timestamptz '2027-01-01' + i * interval '1 minute'
FROM generate_series(1, 500000) AS i;
ANALYZE;
EXPLAIN ANALYZE SELECT id, status FROM applications WHERE organization_id = 123;
CREATE INDEX applications_organization_id_idx ON applications (organization_id);
EXPLAIN ANALYZE SELECT id, status FROM applications WHERE organization_id = 123;
EXPLAIN ANALYZE SELECT id FROM applications WHERE event_id = 7 AND status = 'approved' ORDER BY created_at DESC LIMIT 20;
CREATE INDEX applications_event_status_created_idx ON applications (event_id, status, created_at);
EXPLAIN ANALYZE SELECT id FROM applications WHERE event_id = 7 AND status = 'approved' ORDER BY created_at DESC LIMIT 20;
EXPLAIN ANALYZE SELECT id FROM applications WHERE status = 'approved' AND created_at > '2027-06-01' LIMIT 20;
EXPLAIN ANALYZE SELECT id FROM applications ORDER BY id LIMIT 20 OFFSET 400000;
EXPLAIN ANALYZE SELECT id FROM applications WHERE id > 400000 ORDER BY id LIMIT 20;
