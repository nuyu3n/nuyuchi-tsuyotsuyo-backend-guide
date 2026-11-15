CREATE TABLE applications (
  id      bigint PRIMARY KEY,
  title   text   NOT NULL,
  status  text   NOT NULL,
  version int    NOT NULL DEFAULT 1
);
INSERT INTO applications (id, title, status) VALUES (42, 'お化け屋敷', 'draft');
-- Aさん、Bさんとも version = 1 を読んだ
UPDATE applications SET title = '本格お化け屋敷', version = version + 1 WHERE id = 42 AND version = 1;
UPDATE applications SET title = 'ゆるいお化け屋敷', version = version + 1 WHERE id = 42 AND version = 1;
SELECT * FROM applications;
UPDATE applications SET status = 'submitted' WHERE id = 42 AND status = 'draft';
UPDATE applications SET status = 'submitted' WHERE id = 42 AND status = 'draft';
SHOW transaction_isolation;
