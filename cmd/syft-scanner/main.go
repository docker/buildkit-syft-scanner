package main

import (
	"fmt"
	"os"

	"github.com/anchore/go-logger"
	alogrus "github.com/anchore/go-logger/adapter/logrus"
	"github.com/anchore/stereoscope"
	"github.com/anchore/syft/syft"
	"github.com/docker/buildkit-syft-scanner/internal"
	"github.com/docker/buildkit-syft-scanner/version"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := enableLogs(); err != nil {
		panic(fmt.Sprintf("unable to initialize logger: %+v", err))
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
