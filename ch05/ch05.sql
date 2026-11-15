CREATE TABLE users (
  id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  public_id  uuid NOT NULL DEFAULT uuidv7() UNIQUE,
  email      text NOT NULL UNIQUE,
  name       text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE events (
  id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name        text NOT NULL,
  kind        text NOT NULL CHECK (kind IN ('festival', 'sports', 'survey')),
  starts_on   date NOT NULL,
  ends_on     date NOT NULL,
  archived_at timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now(),
  CHECK (starts_on <= ends_on)
);

CREATE TABLE organizations (
  id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_id   bigint NOT NULL REFERENCES events (id),
  name       text NOT NULL,
  kind       text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (event_id, name)
);

CREATE TABLE role_assignments (
  user_id         bigint NOT NULL REFERENCES users (id),
  event_id        bigint NOT NULL REFERENCES events (id),
  role            text   NOT NULL CHECK (role IN ('leader', 'staff', 'viewer')),
  organization_id bigint REFERENCES organizations (id),
  created_at      timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, event_id, role)
);

INSERT INTO users (email, name) VALUES ('a@example.com', 'Aさん'), ('b@example.com', 'Bさん');
INSERT INTO events (name, kind, starts_on, ends_on)
VALUES ('文化祭2027', 'festival', '2027-11-03', '2027-11-04'),
       ('スポーツ大会2027', 'sports', '2027-06-10', '2027-06-10');
INSERT INTO organizations (event_id, name, kind) VALUES (1, '写真部', 'club'), (2, '2年1組', 'class');
INSERT INTO role_assignments (user_id, event_id, role) VALUES (1, 1, 'staff');
INSERT INTO role_assignments (user_id, event_id, role, organization_id) VALUES (1, 2, 'leader', 2);

SELECT id, public_id, email FROM users ORDER BY id;

INSERT INTO users (email, name) VALUES ('a@example.com', 'もう1人のAさん');
INSERT INTO events (name, kind, starts_on, ends_on) VALUES ('合唱祭2027', 'chorus', '2027-12-01', '2027-12-01');
INSERT INTO events (name, kind, starts_on, ends_on) VALUES ('逆さまの行事', 'survey', '2027-12-02', '2027-12-01');
INSERT INTO organizations (event_id, name, kind) VALUES (99, '幽霊部', 'club');
INSERT INTO organizations (event_id, name, kind) VALUES (1, '写真部', 'club');
DELETE FROM events WHERE id = 1;

SELECT u.name, e.name AS event, r.role
FROM role_assignments AS r
JOIN users AS u ON u.id = r.user_id
JOIN events AS e ON e.id = r.event_id
ORDER BY e.id;

CREATE TABLE form_versions (
  id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_id   bigint NOT NULL REFERENCES events (id),
  version    int    NOT NULL,
  fields     jsonb  NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (event_id, version)
);

CREATE TABLE form_answers (
  id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  form_version_id bigint NOT NULL REFERENCES form_versions (id),
  organization_id bigint NOT NULL REFERENCES organizations (id),
  values          jsonb  NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now()
);

INSERT INTO form_versions (event_id, version, fields)
VALUES (1, 1, '[{"key": "title", "label": "企画名", "type": "text", "required": true}]');
INSERT INTO form_answers (form_version_id, organization_id, values)
VALUES (1, 1, '{"title": "写真展"}');

UPDATE events SET archived_at = now() WHERE id = 2;
SELECT name FROM events WHERE archived_at IS NULL;
