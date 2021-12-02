#!/usr/bin/env bash

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOCAL_ROOT_PATH=${SCRIPT_DIR}/../../..

echo -e "${GREEN}Prettifying GraphQL examples...${NC}"
img="prettier:latest"
docker build -t ${img} ${LOCAL_ROOT_PATH}/tests/tools/prettier
docker run --rm -v "${LOCAL_ROOT_PATH}/components/director/examples":/prettier/examples \
     ${img} --write "examples/**/*.graphql"

cd "${LOCAL_ROOT_PATH}/tests/tools/example-index-generator/" || exit
EXAMPLES_DIRECTORY="${LOCAL_ROOT_PATH}/components/director/examples" go run main.go