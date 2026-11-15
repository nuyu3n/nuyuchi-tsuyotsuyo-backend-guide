CREATE TABLE IF NOT EXISTS role_assignments (
  user_id         bigint NOT NULL,
  event_id        bigint NOT NULL,
  role            text   NOT NULL CHECK (role IN ('leader', 'staff', 'viewer')),
  organization_id bigint,
  PRIMARY KEY (user_id, event_id, role),
  CHECK ((role = 'leader') = (organization_id IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS applications (
  id              bigint PRIMARY KEY,
  event_id        bigint NOT NULL,
  organization_id bigint NOT NULL
);
