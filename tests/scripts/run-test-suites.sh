#! /bin/bash -v

docker run -p 8080:8080 -p 8081:8081 -p 3000:3000 -d --name openfga-cli-tests openfga/openfga run

set +e

commander test --dir ./tests

exit_code=$?

docker stop openfga-cli-tests
docker rm openfga-cli-tests

rm -rf tests/fixtures/identifiers

exit $exit_code
