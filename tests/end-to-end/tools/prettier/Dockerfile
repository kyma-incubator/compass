# This image is used in `gen-examples.sh` script to prettify our GraphQL examples

FROM node:10-alpine

RUN npm install --global prettier && npm cache --force clean

WORKDIR /prettier
ENTRYPOINT ["prettier"]