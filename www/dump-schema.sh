#!/bin/bash
/Library/PostgreSQL/9.3/bin/pg_dump -C -f ./sql/db.sql -s -U www oneintenband
