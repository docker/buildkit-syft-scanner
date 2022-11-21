package internal

import (
	"fmt"
	"path/filepath"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/pkg/cataloger"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
)

type Target struct {
	Path string
}

func (t Target) Name() string {
	return filepath.Base(t.Path)
}

func (t Target) Scan() (sbom.SBOM, error) {
	inputSrc := "dir:" + t.Path
	input, err := source.ParseInput(inputSrc, "", false)
	if err != nil {
		return sbom.SBOM{}, fmt.Errorf("failed to parse user input %q: %w", inputSrc, err)
	}

	src, cleanup, err := source.New(*input, nil, nil)
	if err != nil {
		return sbom.SBOM{}, fmt.Errorf("failed to construct source from user input %q: %w", inputSrc, err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	result := sbom.SBOM{
		Source: src.Metadata,
		Descriptor: sbom.Descriptor{
			Name:    "syft",
			Version: SyftVersion,
		},
	}

	packageCatalog, relationships, theDistro, err := syft.CatalogPackages(src, cataloger.DefaultConfig())
	if err != nil {
		return sbom.SBOM{}, err
	}

	result.Artifacts.PackageCatalog = packageCatalog
	result.Artifacts.LinuxDistribution = theDistro
	result.Relationships = relationships

	return result, nil
}

var (
	SyftVersion = "[not provided]"
)
