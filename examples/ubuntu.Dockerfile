# syntax=docker/dockerfile:1.5

FROM ubuntu as base
ARG BUILDKIT_SBOM_SCAN_STAGE=true
RUN apt-get update && apt-get install -y git
COPY <<EOF /empty
EOF

FROM scratch
COPY --from=base /empty /
