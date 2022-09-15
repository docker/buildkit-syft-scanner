#syntax=docker/dockerfile-upstream:master-labs

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.1.2 AS xx

FROM golang as build-base
COPY --link --from=xx / /

FROM build-base as build
ARG SYFT_VERSION=b0fc955e0c406a12d8aaddcd8ececda89cbcddce
ADD https://github.com/anchore/syft.git#${SYFT_VERSION} /syft
WORKDIR /syft
ARG TARGETPLATFORM
ENV CGO_ENABLED=0
RUN \
  --mount=target=/root/.cache,type=cache \
  xx-go build -ldflags '-extldflags -static' -o /usr/bin/syft ./cmd/syft && \
  xx-verify --static /usr/bin/syft

FROM alpine:latest
COPY --from=build /usr/bin/syft /usr/bin/syft

COPY <<-"EOF" /entrypoint.sh
	#!/bin/sh
	set -e
	for src in "${BUILDKIT_SCAN_SOURCES:?}"/*; do
		dest="${BUILDKIT_SCAN_DESTINATIONS:?}"/$(basename "$src")
		echo syft --output spdx-json="$dest/spdx.json" "$src"
		syft --output spdx-json="$dest/spdx.json" "$src"
		cat <<BUNDLE > "$dest/index.json"
		[
		  {
		    "kind": "in-toto",
		    "path": "spdx.json",
		    "in-toto": {
		      "predicate-type": "https://spdx.dev/Document"
		    }
		  }
		]
		BUNDLE
	done
	find "${BUILDKIT_SCAN_DESTINATIONS:?}/"
EOF
CMD sh /entrypoint.sh
