#!/bin/bash
/Library/PostgreSQL/9.3/bin/pg_dump -C -a -f ./sql/data.sql -U band -d band
