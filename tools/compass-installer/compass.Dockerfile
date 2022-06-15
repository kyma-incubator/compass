# The base version of the Kyma Operator that will be used to build Compass Installer
# TODO: delete the /compass/tools dir once kyma2.0 reaches prod
ARG INSTALLER_VERSION="4260fe58"
ARG INSTALLER_DIR=eu.gcr.io/kyma-project
FROM $INSTALLER_DIR/kyma-operator:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /chart /kyma/injected/resources
