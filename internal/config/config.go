package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/engine"
	"gopkg.in/yaml.v3"
)

const CurrentSchemaVersion = 1

type RunConfig struct {
	SchemaVersion int               `json:"schema_version" yaml:"schema_version"`
	Survey        SurveyConfig      `json:"survey" yaml:"survey"`
	Run           RuntimeConfig     `json:"run" yaml:"run"`
	Questions     []QuestionConfig  `json:"questions" yaml:"questions"`
	Proxy         ProxyConfig       `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	ReverseFill   ReverseFillConfig `json:"reverse_fill,omitempty" yaml:"reverse_fill,omitempty"`
	RandomUA      RandomUAConfig    `json:"random_ua,omitempty" yaml:"random_ua,omitempty"`
	Answer        map[string]any    `json:"answer,omitempty" yaml:"answer,omitempty"`
}

type SurveyConfig struct {
	URL      string `json:"url" yaml:"url"`
	Provider string `json:"provider" yaml:"provider"`
}

type RuntimeConfig struct {
	Target           int             `json:"target" yaml:"target"`
	Concurrency      int             `json:"concurrency" yaml:"concurrency"`
	Mode             engine.Mode     `json:"mode" yaml:"mode"`
	FailureThreshold int             `json:"failure_threshold" yaml:"failure_threshold"`
	FailStopEnabled  bool            `json:"fail_stop_enabled" yaml:"fail_stop_enabled"`
	Headless         bool            `json:"headless" yaml:"headless"`
	SubmitInterval   DurationRange   `json:"submit_interval,omitempty" yaml:"submit_interval,omitempty"`
	AnswerDuration   DurationRange   `json:"answer_duration,omitempty" yaml:"answer_duration,omitempty"`
	TimedMode        TimedModeConfig `json:"timed_mode,omitempty" yaml:"timed_mode,omitempty"`
}

type DurationRange struct {
	MinSeconds int `json:"min_seconds" yaml:"min_seconds"`
	MaxSeconds int `json:"max_seconds" yaml:"max_seconds"`
}

type TimedModeConfig struct {
	Enabled                bool `json:"enabled" yaml:"enabled"`
	RefreshIntervalSeconds int  `json:"refresh_interval_seconds" yaml:"refresh_interval_seconds"`
}

type ProxyConfig struct {
	Enabled       bool   `json:"enabled" yaml:"enabled"`
	Source        string `json:"source,omitempty" yaml:"source,omitempty"`
	CustomAPI     string `json:"custom_api,omitempty" yaml:"custom_api,omitempty"`
	AreaCode      string `json:"area_code,omitempty" yaml:"area_code,omitempty"`
	OccupyMinutes int    `json:"occupy_minutes,omitempty" yaml:"occupy_minutes,omitempty"`
}

type ReverseFillConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	SourcePath string `json:"source_path,omitempty" yaml:"source_path,omitempty"`
	Format     string `json:"format,omitempty" yaml:"format,omitempty"`
	StartRow   int    `json:"start_row,omitempty" yaml:"start_row,omitempty"`
}

type RandomUAConfig struct {
	Enabled bool           `json:"enabled" yaml:"enabled"`
	Keys    []string       `json:"keys,omitempty" yaml:"keys,omitempty"`
	Ratios  map[string]int `json:"ratios,omitempty" yaml:"ratios,omitempty"`
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
		Proxy:         DefaultProxyConfig(),
		ReverseFill:   DefaultReverseFillConfig(),
		RandomUA:      DefaultRandomUAConfig(),
		Questions:     []QuestionConfig{},
	}
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		Target:           1,
		Concurrency:      1,
		Mode:             engine.ModeHybrid,
		FailureThreshold: 1,
		FailStopEnabled:  true,
		Headless:         true,
		TimedMode: TimedModeConfig{
			RefreshIntervalSeconds: 3,
		},
	}
}

func DefaultProxyConfig() ProxyConfig {
	return ProxyConfig{
		Source:        "default",
		OccupyMinutes: 1,
	}
}

func DefaultReverseFillConfig() ReverseFillConfig {
	return ReverseFillConfig{
		Format:   "auto",
		StartRow: 1,
	}
}

func DefaultRandomUAConfig() RandomUAConfig {
	return RandomUAConfig{
		Ratios: map[string]int{
			"wechat": 33,
			"mobile": 33,
			"pc":     34,
		},
	}
}

func LoadRunConfig(path string) (RunConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RunConfig{}, fmt.Errorf("read config %q: %w", path, err)
	}

	cfg := DefaultRunConfig()
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
	if err := c.Run.Validate(); err != nil {
		return err
	}
	if err := c.Proxy.Validate(); err != nil {
		return fmt.Errorf("proxy: %w", err)
	}
	if err := c.ReverseFill.Validate(); err != nil {
		return fmt.Errorf("reverse_fill: %w", err)
	}
	if err := c.RandomUA.Validate(); err != nil {
		return fmt.Errorf("random_ua: %w", err)
	}
	return nil
}

func (c RuntimeConfig) Validate() error {
	if c.Target <= 0 {
		return fmt.Errorf("run.target must be greater than 0")
	}
	if c.Concurrency <= 0 {
		return fmt.Errorf("run.concurrency must be greater than 0")
	}
	if c.FailureThreshold < 0 {
		return fmt.Errorf("run.failure_threshold must not be negative")
	}
	mode, err := engine.ParseMode(c.Mode.String())
	if err != nil {
		return fmt.Errorf("run.mode: %w", err)
	}
	if err := engine.ValidateConcurrency(mode, c.Concurrency); err != nil {
		return fmt.Errorf("run.concurrency: %w", err)
	}
	if err := c.SubmitInterval.Validate("run.submit_interval"); err != nil {
		return err
	}
	if err := c.AnswerDuration.Validate("run.answer_duration"); err != nil {
		return err
	}
	if c.TimedMode.RefreshIntervalSeconds < 0 {
		return fmt.Errorf("run.timed_mode.refresh_interval_seconds must not be negative")
	}
	return nil
}

func (r DurationRange) Validate(name string) error {
	if r.MinSeconds < 0 {
		return fmt.Errorf("%s.min_seconds must not be negative", name)
	}
	if r.MaxSeconds < 0 {
		return fmt.Errorf("%s.max_seconds must not be negative", name)
	}
	if r.MaxSeconds > 0 && r.MinSeconds > r.MaxSeconds {
		return fmt.Errorf("%s.min_seconds must be less than or equal to max_seconds", name)
	}
	return nil
}

func (c ProxyConfig) Validate() error {
	source := strings.TrimSpace(c.Source)
	if source == "" {
		source = "default"
	}
	switch source {
	case "default", "custom":
	default:
		return fmt.Errorf("source %q is unsupported", c.Source)
	}
	if c.Enabled && source == "custom" && strings.TrimSpace(c.CustomAPI) == "" {
		return fmt.Errorf("custom_api is required when source is custom")
	}
	if c.OccupyMinutes < 0 {
		return fmt.Errorf("occupy_minutes must not be negative")
	}
	return nil
}

func (c ReverseFillConfig) Validate() error {
	format := strings.TrimSpace(c.Format)
	if format == "" {
		format = "auto"
	}
	switch format {
	case "auto", "wjx_sequence", "wjx_score", "wjx_text":
	default:
		return fmt.Errorf("format %q is unsupported", c.Format)
	}
	if c.Enabled && strings.TrimSpace(c.SourcePath) == "" {
		return fmt.Errorf("source_path is required when enabled")
	}
	if c.StartRow < 0 {
		return fmt.Errorf("start_row must not be negative")
	}
	return nil
}

func (c RandomUAConfig) Validate() error {
	for key, ratio := range c.Ratios {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("ratio key is required")
		}
		if ratio < 0 {
			return fmt.Errorf("ratio %q must not be negative", key)
		}
	}
	return nil
}
