package clio

import (
	"github.com/wagoodman/go-partybus"
)

type UIConstructor func(Config) ([]UI, error)

type UI interface {
	Setup(subscription partybus.Unsubscribable) error
	partybus.Handler
	Teardown(force bool) error
}

var _ UIConstructor = newUI

func newUI(Config) ([]UI, error) {
	// gracefully degrade to no UI if no constructor is configured
	return nil, nil
}
