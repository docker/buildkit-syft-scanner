variable "GO_VERSION" {
  default = "1.19"
}

target "_common" {
  args = {
    GO_VERSION = GO_VERSION
  }
}

# Special target: https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {
  tags = ["buildkit-syft-scanner:local"]
}

group "default" {
  targets = ["image"]
}

target "image" {
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
