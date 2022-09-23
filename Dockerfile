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
	
	env

	scan () {
		echo "Scanning $1"
		out="$(basename $1).spdx.json"
		syft --output spdx-json="${BUILDKIT_SCAN_DESTINATION}/$out" "$1"
		cat <<-BUNDLE >> "${BUILDKIT_SCAN_DESTINATION_INDEX}"
		{
		  "kind": "in-toto",
		  "path": "$out",
		  "in-toto": {
		    "predicate-type": "https://spdx.dev/Document"
		  }
		}
		BUNDLE
	}
	
	scan "$BUILDKIT_SCAN_SOURCE"
	if [ -d "${BUILDKIT_SCAN_SOURCE_EXTRAS:?}" ]; then
		for src in "${BUILDKIT_SCAN_SOURCE_EXTRAS}"/*; do
			scan "$src"
		done
	fi
	
	find "${BUILDKIT_SCAN_DESTINATION:?}/"
EOF
CMD sh /entrypoint.sh
