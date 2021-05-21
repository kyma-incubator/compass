# This image is used in `gen-examples.sh` script to prettify our GraphQL examples

FROM node:10-alpine

RUN npm install --global prettier@2.0.1 && npm cache --force clean

WORKDIR /prettier
ENTRYPOINT ["prettier"]