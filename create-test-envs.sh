#!/bin/sh

rm -rf .test-envs
echo "C go build"
go build -o anonmess-client-cli

echo "C making envs"
mkdir .test-envs

for user in "$@"; do
    mkdir ".test-envs/$user"
    cp anonmess-client-cli ".test-envs/$user/client"
    echo "PROGRAM_DATA_DIR=./.appdata" >> ".test-envs/$user/.env"
done
