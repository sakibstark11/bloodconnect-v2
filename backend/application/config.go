package application

import "time"

type AppConfig struct {
	ProcessRequestWindowDays int
	SearchRadiusKm           float64
	H3HexResolution          int
	WaveSearchInterval       time.Duration
	WaveSearchMaxRetries     int
	WaveSearchRetryDelay     time.Duration
	RequestAcceptanceWindow  time.Duration
	JWTSecret                string
	DefaultPageSize          int
	NotificationPageSize     int
	MinimumDonationWaitDays  int
	JobWorkerTickerInterval  time.Duration
}

func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		ProcessRequestWindowDays: 7,
		SearchRadiusKm:           5.0,
		H3HexResolution:          8,
		WaveSearchInterval:       3 * time.Minute,
		WaveSearchMaxRetries:     3,
		RequestAcceptanceWindow:  1 * time.Hour,
		WaveSearchRetryDelay:     5 * time.Minute,
		JWTSecret:                "super-secret-key-change-me",
		DefaultPageSize:          20,
		NotificationPageSize:     10,
		MinimumDonationWaitDays:  90,
		JobWorkerTickerInterval:  5 * time.Second,
	}
}
