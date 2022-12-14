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
	// HACK: execute the scan inside a chroot, to ensure that symlinks are
	// correctly resolved internally to the mounted image (instead of
	// redirecting to the host).
	//
	// To avoid this, syft needs to support a mode of execution that scans
	// unpacked container filesystems, see https://github.com/anchore/syft/issues/1359.

	var result sbom.SBOM
	err := withChroot(t.Path, func() error {
		inputSrc := "dir:/"
		input, err := source.ParseInput(inputSrc, "", false)
		if err != nil {
			return fmt.Errorf("failed to parse user input %q: %w", inputSrc, err)
		}

		src, cleanup, err := source.New(*input, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to construct source from user input %q: %w", inputSrc, err)
		}
		src.Metadata.Name = t.Name()
		if cleanup != nil {
			defer cleanup()
		}

		result = sbom.SBOM{
			Source: src.Metadata,
			Descriptor: sbom.Descriptor{
				Name:    "syft",
				Version: version.SyftVersion,
			},
		}

		packageCatalog, relationships, theDistro, err := syft.CatalogPackages(src, cataloger.DefaultConfig())
		if err != nil {
			return err
		}

		result.Artifacts.PackageCatalog = packageCatalog
		result.Artifacts.LinuxDistribution = theDistro
		result.Relationships = relationships

		return nil
	})
	if err != nil {
		return sbom.SBOM{}, err
	}

	return result, nil
}
