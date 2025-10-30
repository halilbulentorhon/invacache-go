package option

type DelOptFnc func(*DeleteConfig)

type DeleteConfig struct {
	PublishInvalidation bool
}

func WithDeleteInvalidation() DelOptFnc {
	return func(cfg *DeleteConfig) {
		cfg.PublishInvalidation = true
	}
}

func ApplyDeleteOptions(options []DelOptFnc) DeleteConfig {
	cfg := defaultDeleteConfig()
	for _, opt := range options {
		opt(&cfg)
	}
	return cfg
}

func defaultDeleteConfig() DeleteConfig {
	return DeleteConfig{
		PublishInvalidation: false,
	}
}
