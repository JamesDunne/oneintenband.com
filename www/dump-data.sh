#!/bin/bash
pg_dump -C -a -f ./sql/data.sql -U band -d band
