package option

type ClrOptFnc func(*ClearConfig)

type ClearConfig struct {
	PublishInvalidation bool
}

func WithClearInvalidation() ClrOptFnc {
	return func(cfg *ClearConfig) {
		cfg.PublishInvalidation = true
	}
}

func ApplyClearOptions(options []ClrOptFnc) ClearConfig {
	cfg := defaultClearConfig()
	for _, opt := range options {
		opt(&cfg)
	}
	return cfg
}

func defaultClearConfig() ClearConfig {
	return ClearConfig{
		PublishInvalidation: false,
	}
}
