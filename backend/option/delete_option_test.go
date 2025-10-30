package option

import "testing"

func TestDefaultDeleteConfig(t *testing.T) {
	cfg := defaultDeleteConfig()

	if cfg.PublishInvalidation != false {
		t.Error("expected PublishInvalidation false")
	}
}

func TestWithDeleteInvalidation(t *testing.T) {
	opt := WithDeleteInvalidation()

	cfg := &DeleteConfig{}
	opt(cfg)

	if !cfg.PublishInvalidation {
		t.Error("expected PublishInvalidation true")
	}
}

func TestApplyDeleteOptionsNoOptions(t *testing.T) {
	cfg := ApplyDeleteOptions([]DelOptFnc{})

	if cfg.PublishInvalidation != false {
		t.Error("expected PublishInvalidation false")
	}
}

func TestApplyDeleteOptionsSingleOption(t *testing.T) {
	cfg := ApplyDeleteOptions([]DelOptFnc{WithDeleteInvalidation()})

	if !cfg.PublishInvalidation {
		t.Error("expected PublishInvalidation true")
	}
}

func TestApplyDeleteOptionsMultipleInvalidationCalls(t *testing.T) {
	cfg := ApplyDeleteOptions([]DelOptFnc{
		WithDeleteInvalidation(),
		WithDeleteInvalidation(),
	})

	if !cfg.PublishInvalidation {
		t.Error("expected PublishInvalidation true")
	}
}
