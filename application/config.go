package application

import "time"

// AppConfig holds the system-wide configuration values
type AppConfig struct {
	ProcessRequestWindowDays int
	SearchRadiusKm           float64
	H3HexResolution          int
	JobQueueInterval         time.Duration
}

// DefaultAppConfig returns the default configuration for the application
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		ProcessRequestWindowDays: 7,
		SearchRadiusKm:           5.0,
		H3HexResolution:          8,
		JobQueueInterval:         120 * time.Second,
	}
}
