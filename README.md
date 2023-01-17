# BuildKit Syft scanner

This repo packages the [Syft scanner](https://github.com/anchore/syft) as a
[BuildKit SBOM generator](https://github.com/moby/buildkit/blob/master/docs/attestations/sbom.md)
to include scan results with the output of Docker builds.

The [docker/buildkit-syft-scanner](https://hub.docker.com/r/docker/buildkit-syft-scanner)
image implements the BuildKit SBOM scanner protocol defined
[here](https://github.com/moby/buildkit/blob/master/docs/attestations/sbom-protocol.md).

## Usage

To scan an image during build with [buildctl](https://github.com/moby/buildkit):

    $ buildctl build ... \
        --output type=image,name=<image>,push=true \
        --opt attest:sbom=generator=docker/buildkit-syft-scanner

## Development

`buildkit-syft-scanner` uses bake to build the project.

To setup a development environment by cloning the git repository:

    $ git clone https://github.com/docker/buildkit-syft-scanner.git
    $ cd buildkit-syft-scanner

It's recommended to setup an ephemeral local registry to push the development
image to:

    $ docker run -d -p 5000:5000 --rm --name registry registry:2

To build the development image, and push it to `localhost:5000/buildkit-syft-scanner:dev`:

    $ make dev IMAGE=localhost:5000/buildkit-syft-scanner:dev 

To scan an image during build with [buildctl](https://github.com/moby/buildkit)
using the development image:

    $ buildctl build ... \
        --output type=image,name=<image>,push=true \
        --opt attest:sbom=generator=localhost:5000/buildkit-syft-scanner:dev

## Contributing

Want to contribute? Awesome!

`buildkit-syft-scanner` is mostly glue between [BuildKit](https://github.com/moby/buildkit)
and [Syft](https://github.com/anchore/syft), so contributions will mostly
likely belong in one of those projects. This project is intended to be as thin
a compatibility layer as possible, so we have a strong preference for as little
code here as possible.
