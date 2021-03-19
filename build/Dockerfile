# Build the mattermost apps cloud deployer
ARG DOCKER_BUILD_IMAGE=golang:1.16.2
ARG DOCKER_BASE_IMAGE=alpine:3.12.4

FROM ${DOCKER_BUILD_IMAGE} AS build
WORKDIR /mattermost-apps-cloud-deployer/
COPY . /mattermost-apps-cloud-deployer/
RUN apt-get update -yq && apt-get install -yq unzip
RUN make get-terraform
RUN make build


# Final Image
FROM ${DOCKER_BASE_IMAGE}
LABEL name="Mattermost Apps Cloud Deployer" \
  maintainer="cloud-team@mattermost.com" \
  vendor="Mattermost" \
  distribution-scope="public" \
  architecture="x86_64" \
  url="https://mattermost.com" \
  io.k8s.description="Mattermost Apps Cloud Deployer is responsible for Mattermost apps deployment" \
  io.k8s.display-name="Mattermost Apps Cloud Deployer"

ENV CLOUD=/mattermost-apps-cloud-deployer/main \
    USER_UID=10001 \
    USER_NAME=cloud

RUN  apk update && apk add libc6-compat && apk add ca-certificates
COPY --from=build /mattermost-apps-cloud-deployer/build/terraform /usr/local/bin/
COPY --from=build /mattermost-apps-cloud-deployer/build/_output/bin/main /mattermost-apps-cloud-deployer/main
COPY --from=build /mattermost-apps-cloud-deployer/build/bin /usr/local/bin
COPY --from=build /mattermost-apps-cloud-deployer/terraform /mattermost-apps-cloud-deployer/terraform
RUN mkdir /mattermost-apps-cloud-deployer/tmp

RUN  /usr/local/bin/user_setup
WORKDIR /mattermost-apps-cloud-deployer/

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
