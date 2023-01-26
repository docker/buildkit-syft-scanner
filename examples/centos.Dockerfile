# syntax=docker/dockerfile:1.5

FROM centos:7
ARG BUILDKIT_SBOM_SCAN_STAGE=true
RUN yum install -y findutils
COPY <<EOF /empty
EOF

FROM scratch
COPY --from=0 /empty /
