-- /home/krylon/go/src/github.com/blicero/uptimed/database/testdata/overview.sql
-- created on 05. 06. 2023 by Benjamin Walkenhorst
-- (c) 2023 Benjamin Walkenhorst
-- Use at your own risk!

WITH data AS (
SELECT
        h.name,
        row_number() OVER (PARTITION BY h.name ORDER BY r.timestamp DESC) AS hid,
        datetime(r.timestamp, 'unixepoch', 'localtime') AS timestamp,
        r.uptime,
        r.load1,
        r.load5,
        r.load15
FROM host h
RIGHT JOIN record r ON h.id = r.host_id
)
SELECT name, timestamp, uptime, load1, load5, load15
FROM data
WHERE hid = 1
;
