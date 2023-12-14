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

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/anchore/go-logger"
	alogrus "github.com/anchore/go-logger/adapter/logrus"
	"github.com/anchore/stereoscope"
	"github.com/anchore/syft/syft"
	"github.com/docker/buildkit-syft-scanner/internal"
	"github.com/docker/buildkit-syft-scanner/version"
	"github.com/sirupsen/logrus"

	// register sqlite driver for RPMDBs scan support with syft
	_ "modernc.org/sqlite"
)

func main() {
	if err := enableLogs(); err != nil {
		panic(fmt.Sprintf("unable to initialize logger: %+v", err))
	}

	// HACK: ensure that /tmp exists, as syft will fail if it does not
	if err := os.Mkdir("/tmp", 0o777); err != nil && !errors.Is(err, os.ErrExist) {
		panic("could not create /tmp directory")
	}

	logrus.Infof("starting syft scanner for buildkit %s", version.Version)

	scanner, err := internal.NewScannerFromEnvironment()
	if err != nil {
		panic(err)
	}
	if err := scanner.Scan(); err != nil {
		panic(err)
	}
}

const (
	envLogLevel = "LOG_LEVEL"
)

func enableLogs() error {
	level, ok := os.LookupEnv(envLogLevel)
	if !ok {
		level = "warn"
	}

	cfg := alogrus.Config{
		EnableConsole: true,
		Level:         logger.Level(level),
	}
	logWrapper, err := alogrus.New(cfg)
	if err != nil {
		return err
	}
	syft.SetLogger(logWrapper)
	stereoscope.SetLogger(logWrapper.Nested("from-lib", "stereoscope"))

	return nil
}
