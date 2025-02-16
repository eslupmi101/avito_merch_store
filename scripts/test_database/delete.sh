#!/bin/bash

# Stop and remove the PostgreSQL container
docker stop test_postgres_container

docker rm test_postgres_container

echo "Container 'test_postgres_container' stopped and removed."

docker volume rm pgdata_test

echo "Volume 'pgdata_test' removed."
