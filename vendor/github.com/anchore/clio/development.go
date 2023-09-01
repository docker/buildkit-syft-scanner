package clio

import (
	"fmt"
	"strings"

	"github.com/anchore/fangs"
)

const (
	ProfileCPU        Profile = "cpu"
	ProfileMem        Profile = "mem"
	ProfilingDisabled Profile = "none"
)

type Profile string

type DevelopmentConfig struct {
	Profile Profile `yaml:"profile" json:"profile" mapstructure:"profile"`
}

func (d *DevelopmentConfig) DescribeFields(set fangs.FieldDescriptionSet) {
	set.Add(&d.Profile, fmt.Sprintf("capture resource profiling data (available: [%s])", strings.Join([]string{string(ProfileCPU), string(ProfileMem)}, ", ")))
}

func (d *DevelopmentConfig) PostLoad() error {
	p := parseProfile(string(d.Profile))
	if p == "" {
		return fmt.Errorf("invalid profile: %q", d.Profile)
	}
	d.Profile = p
	return nil
}

func parseProfile(profile string) Profile {
	switch strings.ToLower(profile) {
	case "cpu":
		return ProfileCPU
	case "mem", "memory":
		return ProfileMem
	case "none", "", "disabled":
		return ProfilingDisabled
	default:
		return ""
	}
}
