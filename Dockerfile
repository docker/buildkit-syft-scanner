FROM anchore/syft:latest as syft

FROM alpine:latest
COPY --from=syft /syft /bin/syft
CMD syft -o spdx-json="${BUILDKIT_SCAN_DESTINATION:?}" "${BUILDKIT_SCAN_SOURCE:-/}"

