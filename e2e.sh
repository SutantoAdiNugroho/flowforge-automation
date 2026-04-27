#!/bin/bash
set -e

echo "run unit tests.."
make test

echo -e "\nrun applications.."
make run

echo -e "\nwaiting 60 seconds for services to be ready.."
sleep 60

echo -e "\nrun API tests.."
make test_api

echo -e "\nshutdown docker compose"
make shutdown
