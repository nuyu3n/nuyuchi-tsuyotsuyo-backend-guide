CREATE TABLE events (
  id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name       text NOT NULL,
  starts_on  date NOT NULL
);

INSERT INTO events (name, starts_on)
VALUES ('文化祭2027', '2027-11-03'),
       ('スポーツ大会2027', '2027-06-10'),
       ('新入生アンケート2027', '2027-04-15');

SELECT * FROM events;

SELECT name, starts_on
FROM events
WHERE starts_on >= '2027-06-01'
ORDER BY starts_on;

CREATE TABLE organizations (
  id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE applications (
  id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_id        bigint NOT NULL REFERENCES events (id),
  organization_id bigint NOT NULL REFERENCES organizations (id),
  title           text NOT NULL,
  status          text NOT NULL DEFAULT 'draft',
  note            text
);

INSERT INTO organizations (name)
VALUES ('2年1組'), ('写真部'), ('軽音部');

INSERT INTO applications (event_id, organization_id, title, status, note)
VALUES (1, 1, 'お化け屋敷', 'submitted', '暗幕を使う'),
       (1, 2, '写真展', 'approved', NULL),
       (2, 1, 'リレー', 'draft', NULL);

UPDATE applications
SET status = 'approved'
WHERE id = 1;

SELECT id, title, status FROM applications ORDER BY id;

SELECT a.title, o.name AS organization
FROM applications AS a
JOIN organizations AS o ON o.id = a.organization_id
WHERE a.event_id = 1
ORDER BY a.id;

SELECT o.name, a.title
FROM organizations AS o
LEFT JOIN applications AS a
  ON a.organization_id = o.id AND a.event_id = 1
ORDER BY o.id;

SELECT e.name, count(a.id) AS applications
FROM events AS e
LEFT JOIN applications AS a ON a.event_id = e.id
GROUP BY e.id, e.name
ORDER BY e.id;

SELECT status, count(*) FROM applications GROUP BY status HAVING count(*) >= 2;

SELECT count(*) AS by_eq   FROM applications WHERE note = NULL;
SELECT count(*) AS by_is   FROM applications WHERE note IS NULL;
SELECT count(*) AS not_eq  FROM applications WHERE note <> '暗幕を使う';
SELECT count(*) AS all_rows, count(note) AS with_note FROM applications;

CREATE TABLE answers (
  id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  body jsonb NOT NULL
);

INSERT INTO answers (body)
VALUES ('{"grade": 1, "club": "写真部", "interests": ["文化祭", "写真"]}'),
       ('{"grade": 2, "club": null}');

SELECT body ->> 'club' AS club, (body ->> 'grade')::int AS grade FROM answers;
SELECT id FROM answers WHERE body @> '{"grade": 1}';

BEGIN;
DELETE FROM applications;
SELECT count(*) FROM applications;
ROLLBACK;
SELECT count(*) FROM applications;
