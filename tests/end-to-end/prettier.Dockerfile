FROM node:8-alpine

RUN npm install --global prettier && npm cache --force clean

WORKDIR /prettier
ENTRYPOINT ["prettier"]