#!/bin/bash
pg_dump -C -s -f ./sql/db.sql -U band -d band
