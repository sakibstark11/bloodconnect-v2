package application

import "time"

type AppConfig struct {
	ProcessRequestWindowDays int
	SearchRadiusKm           float64
	H3HexResolution          int
	JobQueueInterval         time.Duration
	WaveSearchMaxRetries     int
	WaveSearchRetryDelay     time.Duration
	JWTSecret                string
	DefaultPageSize          int
}

// DefaultAppConfig returns the default configuration for the application
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		ProcessRequestWindowDays: 7,
		SearchRadiusKm:           5.0,
		H3HexResolution:          8,
		JobQueueInterval:         120 * time.Second,
		WaveSearchMaxRetries:     3,
		WaveSearchRetryDelay:     5 * time.Minute,
		JWTSecret:                "super-secret-key-change-me",
		DefaultPageSize:          20,
	}
}
