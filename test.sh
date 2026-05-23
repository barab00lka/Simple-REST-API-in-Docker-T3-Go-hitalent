#!/bin/sh

docker run --name test-db -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test-db -p 5432:5432 -d postgres:alpine

docker build --target test -t t3-app-go-test .
