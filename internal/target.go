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
	"context"
	"fmt"
	"path/filepath"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cataloging/pkgcataloging"
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

func (t Target) Scan(ctx context.Context) (sbom.SBOM, error) {
	src, err := syft.GetSource(context.Background(), t.Path,
		syft.DefaultGetSourceConfig().
			WithBasePath(t.Path).
			WithAlias(source.Alias{Name: t.Name()}))
	if err != nil {
		return sbom.SBOM{}, fmt.Errorf("failed to get source from %q: %w", t.Path, err)
	}

	result, err := syft.CreateSBOM(
		ctx,
		src,
		syft.DefaultCreateSBOMConfig().
			WithCatalogerSelection(
				pkgcataloging.NewSelectionRequest().
					WithDefaults(pkgcataloging.ImageTag).
					WithAdditions("sbom-cataloger")))
	if err != nil {
		return sbom.SBOM{}, err
	}

	result.Descriptor.Name = "syft"
	result.Descriptor.Version = version.SyftVersion
	return *result, nil
}
