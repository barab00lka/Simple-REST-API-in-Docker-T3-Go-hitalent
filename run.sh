#!/bin/sh

WITH_TESTS=false
for arg in "$@"; do
    if [ "$arg" = "--with-tests" ]; then
        WITH_TESTS=true
        break
    fi
done

if [ "$WITH_TESTS" = true ]; then
    docker build -f Dockerfile --target test -t t3-app-go .
else
    docker build -f Dockerfile -t t3-app-go .
fi

docker compose up -d

if [ "$WITH_TESTS" = true ]; then
	docker exec t3-go_server_1 go test -v ./... 
fi
