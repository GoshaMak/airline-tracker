FROM eclipse-temurin:21-jre-jammy AS java
FROM docker:28.5.2-cli AS docker
FROM golang:1.26.5

COPY --from=java /opt/java/openjdk /opt/java/openjdk
COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker
COPY --from=docker /usr/local/libexec/docker/cli-plugins/docker-compose /usr/local/lib/docker/cli-plugins/docker-compose
ENV JAVA_HOME=/opt/java/openjdk
ENV PATH=/opt/java/openjdk/bin:$PATH
ENV JAVA_TOOL_OPTIONS=-Xmx256m

RUN apt-get update && apt-get install -y --no-install-recommends python3 make curl git ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY backend/scripts/install-allure.sh /opt/ci/
RUN allure_bin=$(sh /opt/ci/install-allure.sh /opt) && ln -s "$allure_bin" /usr/local/bin/allure
COPY tests/ci/checkout.sh tests/ci/report.py /opt/ci/
