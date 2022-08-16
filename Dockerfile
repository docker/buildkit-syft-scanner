#syntax=docker/dockerfile-upstream:master-labs

FROM --platform=$BUILDPLATFORM tonistiigi/xx:latest AS xx

FROM golang as build-base
COPY --link --from=xx / /

FROM build-base as build
ARG SYFT_VERSION=main
ADD https://github.com/anchore/syft.git#${SYFT_VERSION} /syft
WORKDIR /syft
ENV CGO_ENABLED=0
RUN \
  --mount=target=/root/.cache,type=cache \
  xx-go build -ldflags '-extldflags -static' -o /usr/bin/syft ./cmd/syft && \
  xx-verify --static /usr/bin/syft

FROM alpine:latest
COPY --from=build /usr/bin/syft /usr/bin/syft

COPY <<-"EOF" /entrypoint.sh
	#!/bin/sh
	for source in "${BUILDKIT_SCAN_SOURCES:?}"/*; do
		name=$(basename "$source")
		syft \
			--output spdx-json="${BUILDKIT_SCAN_DESTINATIONS:?}/$name/spdx.json" \
			"$source"
	done
EOF
CMD sh /entrypoint.sh
