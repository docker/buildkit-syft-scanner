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
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/formats/spdxjson"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/pkg/errors"
)

type Scanner struct {
	Core        Target
	Extras      []Target
	Destination string
}

func (s Scanner) Scan() error {
	for _, target := range append([]Target{s.Core}, s.Extras...) {
		result, err := target.Scan()
		if err != nil {
			return err
		}

		output, err := syft.Encode(result, spdxjson.Format())
		if err != nil {
			return err
		}
		stmt := intoto.Statement{
			StatementHeader: intoto.StatementHeader{
				Type:          intoto.StatementInTotoV01,
				PredicateType: intoto.PredicateSPDX,
			},
			Predicate: json.RawMessage(output),
		}

		outputPath := filepath.Join(s.Destination, target.Name()+".spdx.json")
		f, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		if err := json.NewEncoder(f).Encode(stmt); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

const (
	envScanDestination  = "BUILDKIT_SCAN_DESTINATION"
	envScanSource       = "BUILDKIT_SCAN_SOURCE"
	envScanSourceExtras = "BUILDKIT_SCAN_SOURCE_EXTRAS"
)

func NewScannerFromEnvironment() (*Scanner, error) {
	destPath, err := loadPathFromEnvironment(envScanDestination, true)
	if err != nil {
		return nil, err
	}

	corePath, err := loadPathFromEnvironment(envScanSource, true)
	if err != nil {
		return nil, err
	}
	core := Target{Path: corePath}

	extrasPath, err := loadPathFromEnvironment(envScanSourceExtras, false)
	if err != nil {
		return nil, err
	}
	var extras []Target
	if extrasPath != "" {
		entries, err := os.ReadDir(extrasPath)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			extras = append(extras, Target{
				Path: filepath.Join(extrasPath, entry.Name()),
			})
		}
	}

	scanner := Scanner{
		Destination: destPath,
		Core:        core,
		Extras:      extras,
	}
	return &scanner, nil
}

func loadPathFromEnvironment(name string, required bool) (string, error) {
	p, ok := os.LookupEnv(name)
	if !ok {
		if !required {
			return "", nil
		}
		return "", errors.Errorf("required variable %q not set", name)
	}

	if _, err := os.Stat(p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errors.Wrapf(err, "variable %q (%q) does not exist", name, p)
		}
		return "", err
	}
	return p, nil
}
