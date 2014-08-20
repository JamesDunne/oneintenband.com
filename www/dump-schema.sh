#!/bin/bash
/Library/PostgreSQL/9.3/bin/pg_dump -C -s -f ./sql/db.sql -U band -d band
