#!/bin/bash
set -e

echo "Removing database and web server containers..."
docker-compose down
