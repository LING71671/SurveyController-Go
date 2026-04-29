package config

import "github.com/hungryM0/SurveyController-go/internal/engine"

type RuntimeConfig struct {
	Engine engine.Mode `json:"engine"`
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		Engine: engine.ModeHybrid,
	}
}

func (c RuntimeConfig) Validate() error {
	_, err := engine.ParseMode(c.Engine.String())
	return err
}
