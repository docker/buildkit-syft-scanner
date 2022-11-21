# BuildKit Syft scanner

This repo packages the [Syft scanner](https://github.com/anchore/syft) as a
[BuildKit SBOM generator](https://github.com/moby/buildkit/pull/2983) to
include scan results with the output of Docker builds.

## Usage

To scan an image during build using [buildctl](https://github.com/moby/buildkit):

    $ buildctl build ... \
        --output type=image,name=<image>,push=true --opt attest:sbom=generator=docker/buildkit-syft-scanner
