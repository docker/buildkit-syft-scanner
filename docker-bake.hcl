group "default" {
    targets = ["buildkit-syft-scanner"]
}

target "buildkit-syft-scanner" {
    context = "."
    dockerfile = "Dockerfile"

	tags = ["jedevc/buildkit-syft-scanner:latest"]
    platforms = ["linux/amd64"]
}