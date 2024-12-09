// Copyright 2022 buildkit-syft-scanner authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

variable "GO_VERSION" {
  default = "1.23"
}

# GITHUB_REF is the actual ref that triggers the workflow and used as version
# when tag is pushed: https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
variable "GITHUB_REF" {
  default = ""
}

variable "IMAGE_LOCAL" {
  default = "buildkit-syft-scanner:local"
}

target "_common" {
  args = {
    GO_VERSION = GO_VERSION
    GIT_REF = GITHUB_REF
  }
}

# Special target: https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {
  tags = ["buildkit-syft-scanner:local"]
}

group "default" {
  targets = ["image-local"]
}

target "image-local" {
  inherits = ["_common"]
  tags = ["${IMAGE_LOCAL}"]
  output = ["type=image"]
}

target "image-all" {
  inherits = ["_common", "docker-metadata-action"]
  platforms = [
    "linux/amd64",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le",
    "linux/riscv64",
    "linux/s390x"
  ]
}

target "validate-license" {
  inherits = ["_common"]
  dockerfile = "./hack/dockerfiles/license.Dockerfile"
  target = "validate"
  output = ["type=cacheonly"]
}

target "update-license" {
  inherits = ["_common"]
  dockerfile = "./hack/dockerfiles/license.Dockerfile"
  target = "update"
  output = ["."]
}

target "validate-vendor" {
  inherits = ["_common"]
  dockerfile = "./hack/dockerfiles/vendor.Dockerfile"
  target = "validate"
  output = ["type=cacheonly"]
}

target "update-vendor" {
  inherits = ["_common"]
  dockerfile = "./hack/dockerfiles/vendor.Dockerfile"
  target = "update"
  output = ["."]
}