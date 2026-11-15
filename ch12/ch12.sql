CREATE TABLE users (
  id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  email      text,
  name       text NOT NULL,
  phone      text,
  anonymized_at timestamptz
);
CREATE TABLE events (id bigint PRIMARY KEY, name text NOT NULL, ended_on date NOT NULL);
CREATE TABLE participations (user_id bigint REFERENCES users (id), event_id bigint REFERENCES events (id));
INSERT INTO users (email, name, phone) VALUES ('tanaka@example.com', '田中', '090-0000-0000'), ('suzuki@example.com', '鈴木', NULL);
INSERT INTO events VALUES (1, '文化祭2024', '2024-11-04'), (2, '文化祭2027', '2027-11-04');
INSERT INTO participations VALUES (1, 1), (2, 2);
UPDATE users AS u
SET email = NULL, name = '匿名', phone = NULL, anonymized_at = now()
WHERE u.anonymized_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM participations AS p
    JOIN events AS e ON e.id = p.event_id
    WHERE p.user_id = u.id AND e.ended_on > current_date - interval '1 year'
  );
SELECT id, email, name, phone, anonymized_at IS NOT NULL AS anonymized FROM users ORDER BY id;
SELECT e.name, count(*) FROM participations p JOIN events e ON e.id = p.event_id GROUP BY e.name ORDER BY e.name;
