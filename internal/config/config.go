package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/engine"
	"gopkg.in/yaml.v3"
)

const CurrentSchemaVersion = 1

type RunConfig struct {
	SchemaVersion int              `json:"schema_version" yaml:"schema_version"`
	Survey        SurveyConfig     `json:"survey" yaml:"survey"`
	Run           RuntimeConfig    `json:"run" yaml:"run"`
	Questions     []QuestionConfig `json:"questions" yaml:"questions"`
	Proxy         map[string]any   `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	Answer        map[string]any   `json:"answer,omitempty" yaml:"answer,omitempty"`
}

type SurveyConfig struct {
	URL      string `json:"url" yaml:"url"`
	Provider string `json:"provider" yaml:"provider"`
}

type RuntimeConfig struct {
	Target      int         `json:"target" yaml:"target"`
	Concurrency int         `json:"concurrency" yaml:"concurrency"`
	Mode        engine.Mode `json:"mode" yaml:"mode"`
}

type QuestionConfig struct {
	ID       string         `json:"id" yaml:"id"`
	Kind     string         `json:"kind,omitempty" yaml:"kind,omitempty"`
	Required bool           `json:"required,omitempty" yaml:"required,omitempty"`
	Options  map[string]any `json:"options,omitempty" yaml:"options,omitempty"`
}

func DefaultRunConfig() RunConfig {
	return RunConfig{
		SchemaVersion: CurrentSchemaVersion,
		Run:           DefaultRuntimeConfig(),
		Questions:     []QuestionConfig{},
	}
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		Target:      1,
		Concurrency: 1,
		Mode:        engine.ModeHybrid,
	}
}

func LoadRunConfig(path string) (RunConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RunConfig{}, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg RunConfig
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return RunConfig{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	return cfg, nil
}

func ValidateFile(path string) error {
	cfg, err := LoadRunConfig(path)
	if err != nil {
		return err
	}
	return cfg.Validate()
}

func Migrate(cfg RunConfig) (RunConfig, error) {
	if cfg.SchemaVersion == CurrentSchemaVersion {
		return cfg, nil
	}
	return cfg, fmt.Errorf("unsupported schema_version %d, want %d", cfg.SchemaVersion, CurrentSchemaVersion)
}

func (c RunConfig) Validate() error {
	if c.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("unsupported schema_version %d, want %d", c.SchemaVersion, CurrentSchemaVersion)
	}
	if strings.TrimSpace(c.Survey.URL) == "" {
		return fmt.Errorf("survey.url is required")
	}
	return c.Run.Validate()
}

func (c RuntimeConfig) Validate() error {
	if c.Target <= 0 {
		return fmt.Errorf("run.target must be greater than 0")
	}
	if c.Concurrency <= 0 {
		return fmt.Errorf("run.concurrency must be greater than 0")
	}
	if _, err := engine.ParseMode(c.Mode.String()); err != nil {
		return fmt.Errorf("run.mode: %w", err)
	}
	return nil
}
