package option

import "time"

type OptFnc func(*SetConfig)

type SetConfig struct {
	TTL time.Duration
}

func WithTTL(ttl time.Duration) OptFnc {
	return func(cfg *SetConfig) {
		cfg.TTL = ttl
	}
}

func WithNoExpiration() OptFnc {
	return func(cfg *SetConfig) {
		cfg.TTL = 0
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
		TTL: 0,
	}
}
