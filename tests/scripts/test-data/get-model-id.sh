#! /bin/bash

FILE=./tests/fixtures/identifiers/model-id
if [ -f "$FILE" ]; then
    cat $FILE
    exit 0
fi

STORE_ID=""
STORE_FILE=./tests/fixtures/identifiers/store-id
if [ -f "$STORE_FILE" ]; then
    STORE_ID=$( cat $STORE_FILE )
else
    echo "no store created, must create a store before a model"
    exit 1
fi

model=$( fga model write --file=./tests/fixtures/basic-model.fga --store-id="$STORE_ID" )

mkdir -p ./tests/fixtures/identifiers
echo "$model" | jq -r ".authorization_model_id" > $FILE
cat $FILE