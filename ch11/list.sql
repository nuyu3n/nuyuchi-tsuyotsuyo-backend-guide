CREATE TABLE role_assignments (
  user_id  bigint NOT NULL,
  event_id bigint NOT NULL,
  role     text   NOT NULL CHECK (role IN ('leader', 'staff', 'viewer')),
  organization_id   bigint,
  CHECK ((role = 'leader') = (organization_id IS NOT NULL))
);
CREATE TABLE applications (
  id       bigint PRIMARY KEY,
  event_id bigint NOT NULL,
  organization_id   bigint NOT NULL
);
INSERT INTO role_assignments VALUES (100, 1, 'leader', 10), (200, 1, 'staff', NULL);
INSERT INTO applications VALUES (1, 1, 10), (2, 1, 20), (3, 2, 10);

PREPARE list_applications(bigint, bigint) AS
SELECT a.id, a.organization_id
FROM applications AS a
WHERE a.event_id = $1
  AND EXISTS (
    SELECT 1
    FROM role_assignments AS r
    WHERE r.user_id = $2
      AND r.event_id = a.event_id
      AND (
        r.role IN ('staff', 'viewer')
        OR (r.role = 'leader' AND r.organization_id = a.organization_id)
      )
  )
ORDER BY a.id;

EXECUTE list_applications(1, 100);
EXECUTE list_applications(1, 200);
EXECUTE list_applications(1, 999);
-- 団体のない代表は、CHECK 制約で拒否される（エラーになるのが正しい）
INSERT INTO role_assignments VALUES (300, 1, 'leader', NULL);
