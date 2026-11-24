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
	"os"
	"path/filepath"
	"strings"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cataloging/filecataloging"
	"github.com/anchore/syft/syft/cataloging/pkgcataloging"
	"github.com/anchore/syft/syft/file"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	"github.com/docker/buildkit-syft-scanner/version"
	"github.com/pkg/errors"
)

const (
	envScanSelectCatalogers = "BUILDKIT_SCAN_SELECT_CATALOGERS"
	envScanFileMetadata     = "BUILDKIT_SCAN_FILE_METADATA"
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
		return sbom.SBOM{}, errors.Wrapf(err, "failed to get source from %q", t.Path)
	}

	sr := pkgcataloging.NewSelectionRequest().
		WithDefaults(
			pkgcataloging.ImageTag,
			filecataloging.FileTag, // https://github.com/anchore/syft/pull/3505
		).
		WithAdditions(
			"sbom-cataloger",
		)

	if v, ok := os.LookupEnv(envScanSelectCatalogers); ok {
		sr = pkgcataloging.NewSelectionRequest().WithExpression(strings.Split(v, ",")...)
	}

	cfg := syft.DefaultCreateSBOMConfig().WithCatalogerSelection(sr)
	if v, ok := os.LookupEnv(envScanFileMetadata); ok {
		switch selection := file.Selection(v); selection {
		case file.NoFilesSelection:
			cfg = cfg.WithoutFiles()
		case file.FilesOwnedByPackageSelection, file.AllFilesSelection:
			cfg = cfg.WithFilesConfig(cfg.Files.WithSelection(selection))
		default:
			return sbom.SBOM{}, errors.Errorf("invalid %s value %q: expected %q, %q, or %q",
				envScanFileMetadata, v,
				file.NoFilesSelection,
				file.FilesOwnedByPackageSelection,
				file.AllFilesSelection)
		}
	}

	result, err := syft.CreateSBOM(ctx, src, cfg)
	if err != nil {
		return sbom.SBOM{}, err
	}

	result.Descriptor.Name = "syft"
	result.Descriptor.Version = version.SyftVersion
	return *result, nil
}
