package config

import "time"

// GeneralConfig store model configuration for viper
type GeneralConfig struct {
	BXAPIUrl        string
	LineAccessToken string
	HTTPTimeout     time.Duration
	Delay           time.Duration
}
