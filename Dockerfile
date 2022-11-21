#syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.1.2 AS xx

FROM --platform=$BUILDPLATFORM golang:alpine as build-base
COPY --link --from=xx / /
ENV CGO_ENABLED=0

FROM build-base as build
ARG TARGETPLATFORM
WORKDIR /src
RUN \
  --mount=type=bind,target=. \
  --mount=type=cache,target=/root/.cache \
  xx-go build -ldflags '-extldflags -static' -o /usr/local/bin/syft-scanner ./cmd/syft-scanner && \
  xx-verify --static /usr/local/bin/syft-scanner

FROM scratch
COPY --from=build /usr/local/bin/syft-scanner /bin/syft-scanner
ENV LOG_LEVEL="warn"
ENTRYPOINT [ "/bin/syft-scanner" ]
