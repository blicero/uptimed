-- /home/krylon/go/src/github.com/blicero/uptimed/database/testdata/by_host.sql
-- created on 05. 06. 2023 by Benjamin Walkenhorst
-- (c) 2023 Benjamin Walkenhorst
-- Use at your own risk!

SELECT
  id,
  timestamp,
  uptime,
  load1,
  load5,
  load15
FROM record
WHERE host_id = 1
ORDER BY timestamp
