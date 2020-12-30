FROM maven:3.6.3-jdk-8-slim AS builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/ord-service
ENV COMPONENT_DIR /go/src/github.com/kyma-incubator/compass/components/ord-service
WORKDIR ${BASE_APP_DIR}

RUN apt-get update && apt-get install -y git

#
# Copy files
#

COPY . .

#
# Build app
#

RUN ./scripts/install_dependencies.sh && ./scripts/build.sh -DskipTests && \
    VERSION=$(mvn org.apache.maven.plugins:maven-help-plugin:3.2.0:evaluate -Dexpression=project.version -DforceStdout -q) && \
    mkdir /app && mv ./target/ord-service-$VERSION.jar /app/ord-service.jar

FROM openjdk:8
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

#
# Copy binary
#

COPY --from=builder /app/ord-service.jar /app/

#
# Install curl & CA certs
#

RUN apt-get install -y curl ca-certificates

#
# Run app
#

CMD ["java", "-jar", "ord-service.jar"]
