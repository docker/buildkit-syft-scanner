# syntax=docker/dockerfile:1.5

FROM alpine AS base
ARG BUILDKIT_SBOM_SCAN_STAGE=true
RUN apk add git
COPY <<EOF /empty
EOF

FROM scratch
COPY --from=base /empty /
