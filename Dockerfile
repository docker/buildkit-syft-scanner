#syntax=docker/dockerfile:1.4

FROM anchore/syft:latest as syft

FROM alpine:latest
COPY --from=syft /syft /bin/syft

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
