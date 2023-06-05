-- /home/krylon/go/src/github.com/blicero/uptimed/database/testdata/period.sql
-- created on 05. 06. 2023 by Benjamin Walkenhorst
-- (c) 2023 Benjamin Walkenhorst
-- Use at your own risk!

SELECT
  r.id,
  h.name,
  r.timestamp,
  r.uptime,
  r.load1,
  r.load5,
  r.load15
FROM record r
JOIN host h ON (r.host_id = h.id)
WHERE r.timestamp BETWEEN 0 AND (1<<60)
ORDER BY r.timestamp
