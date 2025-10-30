package option

import "time"

type OptFnc func(*SetConfig)

type SetConfig struct {
	TTL                 time.Duration
	PublishInvalidation bool
	NoExpiration        bool
}

func WithTTL(ttl time.Duration) OptFnc {
	return func(cfg *SetConfig) {
		cfg.TTL = ttl
		cfg.NoExpiration = false
	}
}

func WithInvalidation() OptFnc {
	return func(cfg *SetConfig) {
		cfg.PublishInvalidation = true
	}
}

func WithNoExpiration() OptFnc {
	return func(cfg *SetConfig) {
		cfg.TTL = 0
		cfg.NoExpiration = true
	}
}

func ApplyOptions(options []OptFnc) SetConfig {
	cfg := defaultSetConfig()
	for _, opt := range options {
		opt(&cfg)
	}
	return cfg
}

func defaultSetConfig() SetConfig {
	return SetConfig{
		TTL:                 0,
		PublishInvalidation: false,
		NoExpiration:        false,
	}
}
