#! /bin/bash

FILE=./tests/fixtures/identifiers/store-id
if [ -f "$FILE" ]; then
    cat $FILE
    exit 0
fi

store=$( fga store create --name "integration-test-store" )

mkdir -p ./tests/fixtures/identifiers
echo "$store" | jq -r ".store.id" > $FILE
cat $FILE