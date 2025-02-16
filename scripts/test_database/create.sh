#!/bin/bash

# Create and start the PostgreSQL container
docker run -d \
  --name test_postgres_container \
  -e POSTGRES_USER=test_user \
  -e POSTGRES_PASSWORD=test_password \
  -e POSTGRES_DB=test_db \
  -p 5433:5432 \
  -v pgdata_test:/var/lib/postgresql/data \
  postgres:13-alpine

echo "Container 'test_postgres_container' created and started on port 5433."
