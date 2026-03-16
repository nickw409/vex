package spec

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Spec struct {
	Feature     string     `yaml:"feature"`
	Description string     `yaml:"description"`
	Behaviors   []Behavior `yaml:"behaviors"`
}

type Behavior struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func Load(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spec: %w", err)
	}

	var s Spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing spec: %w", err)
	}

	if err := s.validate(); err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Spec) validate() error {
	if s.Feature == "" {
		return fmt.Errorf("spec missing required field: feature")
	}
	if len(s.Behaviors) == 0 {
		return fmt.Errorf("spec must have at least one behavior")
	}
	for i, b := range s.Behaviors {
		if b.Name == "" {
			return fmt.Errorf("behavior %d missing required field: name", i)
		}
		if b.Description == "" {
			return fmt.Errorf("behavior %q missing required field: description", b.Name)
		}
	}
	return nil
}
