#!/usr/bin/env bash

readonly CI_FLAG=ci

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

echo -e "${INVERTED}"
echo "USER: " + $USER
echo "PATH: " + $PATH
echo "GOPATH:" + $GOPATH
echo -e "${NC}"

##
# DEP ENSURE
##
dep ensure -v --vendor-only
ensureResult=$?
if [ ${ensureResult} != 0 ]; then
	echo -e "${RED}✗ dep ensure -v --vendor-only${NC}\n$ensureResult${NC}"
	exit 1
else echo -e "${GREEN}√ dep ensure -v --vendor-only${NC}"
fi


./gqlgen.sh
genStatus=$?

if [[ ${genStatus} != 0 ]];then
    echo -e "${RED}✗ gqlgen.sh${NC}\n${genStatus}${NC}"
    exit 1
  else
    echo -e "${GREEN}√ gqlgen.sh${NC}"
fi

##
# Ensuring that GraphQL schema and code are in-sync
##
if [[ "$1" == "$CI_FLAG" ]]; then
  if [[ -n $(git status -s) ]]; then
    echo -e "${RED}✗ gqlgen.sh modified some files, schema and code are out-of-sync${NC}"
    exit 1
  fi
fi


##
# GO BUILD
##
buildEnv=""
if [[ "$1" == "$CI_FLAG" ]]; then
	# build binary statically
	buildEnv="env CGO_ENABLED=0 GOOS=linux GOARCH=amd64"
fi

${buildEnv} go build -o bin/provisioner ./cmd/main.go
goBuildResult=$?
rm bin/provisioner

if [ ${goBuildResult} != 0 ]; then
	echo -e "${RED}✗ go build${NC}\n$goBuildResult${NC}"
	exit 1
else echo -e "${GREEN}√ go build${NC}"
fi

##
# DEP STATUS
##
echo "? dep status"
depResult=$(dep status -v)
if [[ $? != 0 ]]
    then
        echo -e "${RED}✗ dep status\n$depResult${NC}"
        exit 1;
    else  echo -e "${GREEN}√ dep status${NC}"
fi

##
# GO TEST
##
TEST_OUTPUT="gotests.log"
echo "? go test"
go test -v ./... > "$TEST_OUTPUT"
# Check if tests passed
if [[ $? != 0 ]]; then
	echo -e "${RED}✗ go test\n${NC}"
	cat "$TEST_OUTPUT"; rm "$TEST_OUTPUT"
	exit 1;
else
	echo -e "${GREEN}√ go test${NC}"
	rm "$TEST_OUTPUT"
fi

filesToCheck=$(find . -type f -name "*.go" | egrep -v "\/vendor\/|_*/automock/|_*/testdata/|/pkg\/|_*export_test.go")
#
# GO IMPORTS
#
go build -o bin/goimports-vendored ./vendor/golang.org/x/tools/cmd/goimports
goImportsResult=$(echo "${filesToCheck}" | xargs -L1 ./bin/goimports-vendored -w -l)
rm bin/goimports-vendored

if [[ $(echo ${#goImportsResult}) != 0 ]]
	then
    	echo -e "${RED}✗ goimports ${NC}\n$goImportsResult${NC}"
    	exit 1;
	else echo -e "${GREEN}√ goimports ${NC}"
fi

#
# GO FMT
#
goFmtResult=$(echo "${filesToCheck}" | xargs -L1 go fmt)
if [[ $(echo ${#goFmtResult}) != 0 ]]
	then
    	echo -e "${RED}✗ go fmt${NC}\n$goFmtResult${NC}"
    	exit 1;
	else echo -e "${GREEN}√ go fmt${NC}"
fi

##
# ERRCHECK
##
go build -o bin/errcheck-vendored ./vendor/github.com/kisielk/errcheck
buildErrCheckResult=$?
if [[ ${buildErrCheckResult} != 0 ]]; then
    echo -e "${RED}✗ go build errcheck${NC}\n${buildErrCheckResult}${NC}"
    exit 1
fi

errCheckResult=$(./bin/errcheck-vendored -blank -asserts -ignoregenerated ./...)
rm bin/errcheck-vendored

if [[ $(echo ${#errCheckResult}) != 0 ]]; then
    echo -e "${RED}✗ [errcheck] unchecked error in:${NC}\n${errCheckResult}${NC}"
    exit 1
else echo -e "${GREEN}√ errcheck ${NC}"
fi

#
# GO VET
#
goVetResult=$(echo "${filesToCheck}" | xargs -L1 go vet)
if [[ $(echo ${#goVetResult}) != 0 ]]
	then
    	echo -e "${RED}✗ go vet${NC}\n$goVetResult${NC}"
    	exit 1;
	else echo -e "${GREEN}√ go vet${NC}"
fi
