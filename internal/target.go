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

package internal

import (
	"fmt"
	"path/filepath"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/pkg/cataloger"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	"github.com/docker/buildkit-syft-scanner/version"
)

type Target struct {
	Path string
}

func (t Target) Name() string {
	return filepath.Base(t.Path)
}

func (t Target) Scan() (sbom.SBOM, error) {
	src, err := source.NewFromDirectoryRootWithName(t.Path, t.Name())
	if err != nil {
		return sbom.SBOM{}, fmt.Errorf("failed to create source from %q: %w", t.Path, err)
	}
	result := sbom.SBOM{
		Source: src.Metadata,
		Descriptor: sbom.Descriptor{
			Name:    "syft",
			Version: version.SyftVersion,
		},
	}

	packageCatalog, relationships, theDistro, err := syft.CatalogPackages(&src, cataloger.DefaultConfig())
	if err != nil {
		return sbom.SBOM{}, err
	}

	result.Artifacts.PackageCatalog = packageCatalog
	result.Artifacts.LinuxDistribution = theDistro
	result.Relationships = relationships

	if err != nil {
		return sbom.SBOM{}, err
	}

	return result, nil
}
