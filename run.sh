#!/bin/sh

docker build -f Dockerfile -t t3-app-go && docker compose up -d
