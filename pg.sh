#!/bin/bash

docker run -p 5432:5432 -e POSTGRES_DB=appclacks -e POSTGRES_USER=appclacks \
       -e POSTGRES_PASSWORD=appclacks \
       timescale/timescaledb:latest-pg16
