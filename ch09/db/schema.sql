CREATE TABLE organizations (
  id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE applications (
  id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_id        bigint NOT NULL,
  organization_id bigint NOT NULL REFERENCES organizations (id),
  title           text   NOT NULL,
  status          text   NOT NULL DEFAULT 'draft',
  note            text,
  created_at      timestamptz NOT NULL DEFAULT now()
);
