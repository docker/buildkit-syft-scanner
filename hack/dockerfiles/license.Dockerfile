# syntax=docker/dockerfile:1

ARG LICENSE_HOLDER="buildkit-syft-scanner authors"
ARG LICENSE_TYPE="apache"
ARG LICENSE_FILES=".*\(Dockerfile\|Makefile\|\.go\|\.hcl\|\.sh\)"
ARG ADDLICENSE_VERSION="v1.0.0"

FROM ghcr.io/google/addlicense:${ADDLICENSE_VERSION} AS addlicense

FROM alpine:3.16 AS base
WORKDIR /src
RUN apk add --no-cache cpio findutils git

FROM base AS set
ARG LICENSE_HOLDER
ARG LICENSE_TYPE
ARG LICENSE_FILES
RUN --mount=type=bind,target=.,rw \
    --mount=from=addlicense,source=/app/addlicense,target=/usr/bin/addlicense \
    find . -regex "${LICENSE_FILES}" -not -path "./vendor/*" | xargs addlicense -c "$LICENSE_HOLDER" -l "$LICENSE_TYPE" && \
    mkdir /out && \
    find . -regex "${LICENSE_FILES}" -not -path "./vendor/*" | cpio -pdm /out

FROM scratch AS update
COPY --from=set /out /

FROM base AS validate
ARG LICENSE_HOLDER
ARG LICENSE_TYPE
ARG LICENSE_FILES
RUN --mount=type=bind,target=. \
    --mount=from=addlicense,source=/app/addlicense,target=/usr/bin/addlicense \
    find . -regex "${LICENSE_FILES}" -not -path "./vendor/*" | xargs addlicense -check -c "$LICENSE_HOLDER" -l "$LICENSE_TYPE"
