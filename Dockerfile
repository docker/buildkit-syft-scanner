# Copyright 2022 buildkit-syft-scanner authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# syntax=docker/dockerfile-upstream:master

ARG GO_VERSION="1.19"
ARG ALPINE_VERSION="3.17"
ARG XX_VERSION="1.1.2"

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
COPY --from=xx / /
ENV CGO_ENABLED=0
RUN apk add --no-cache file git
WORKDIR /src

FROM base AS version
ARG GIT_REF
RUN --mount=target=. <<EOT
  set -e
  case "$GIT_REF" in
    refs/tags/v*) version="${GIT_REF#refs/tags/}" ;;
    *) version=$(git describe --match 'v[0-9]*' --dirty='.m' --always --tags) ;;
  esac
  pkg=github.com/docker/buildkit-syft-scanner
  echo "${version}" | tee /tmp/.version
  echo "-extldflags -static -X ${pkg}/version.Version=${version} -X ${pkg}/version.SyftVersion=$(go list -mod=mod -u -m -f '{{.Version}}' 'github.com/anchore/syft')" | tee /tmp/.ldflags
EOT

FROM base as build
RUN --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download -x
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
    --mount=type=bind,from=version,source=/tmp/.ldflags,target=/tmp/.ldflags \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache <<EOT
  set -e
  xx-go build -trimpath -ldflags "$(cat /tmp/.ldflags)" -o /usr/local/bin/syft-scanner ./cmd/syft-scanner
  xx-verify --static /usr/local/bin/syft-scanner
EOT

FROM scratch
COPY --from=build /usr/local/bin/syft-scanner /bin/syft-scanner
ENV LOG_LEVEL="warn"
ENTRYPOINT [ "/bin/syft-scanner" ]
