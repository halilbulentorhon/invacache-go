package option

import "testing"

func TestDefaultClearConfig(t *testing.T) {
	cfg := defaultClearConfig()

	if cfg.PublishInvalidation != false {
		t.Error("expected PublishInvalidation false")
	}
}

func TestWithClearInvalidation(t *testing.T) {
	opt := WithClearInvalidation()

	cfg := &ClearConfig{}
	opt(cfg)

	if !cfg.PublishInvalidation {
		t.Error("expected PublishInvalidation true")
	}
}

func TestApplyClearOptionsNoOptions(t *testing.T) {
	cfg := ApplyClearOptions([]ClrOptFnc{})

	if cfg.PublishInvalidation != false {
		t.Error("expected PublishInvalidation false")
	}
}

func TestApplyClearOptionsSingleOption(t *testing.T) {
	cfg := ApplyClearOptions([]ClrOptFnc{WithClearInvalidation()})

	if !cfg.PublishInvalidation {
		t.Error("expected PublishInvalidation true")
	}
}

func TestApplyClearOptionsMultipleInvalidationCalls(t *testing.T) {
	cfg := ApplyClearOptions([]ClrOptFnc{
		WithClearInvalidation(),
		WithClearInvalidation(),
	})

	if !cfg.PublishInvalidation {
		t.Error("expected PublishInvalidation true")
	}
}
