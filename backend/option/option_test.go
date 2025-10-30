package option

import (
	"testing"
	"time"
)

func TestDefaultSetConfig(t *testing.T) {
	cfg := defaultSetConfig()

	if cfg.TTL != 0 {
		t.Errorf("expected TTL 0, got %v", cfg.TTL)
	}
	if cfg.PublishInvalidation != false {
		t.Error("expected PublishInvalidation false")
	}
	if cfg.NoExpiration != false {
		t.Error("expected NoExpiration false")
	}
}

func TestWithTTL(t *testing.T) {
	ttl := 5 * time.Minute
	opt := WithTTL(ttl)

	cfg := &SetConfig{}
	opt(cfg)

	if cfg.TTL != ttl {
		t.Errorf("expected TTL %v, got %v", ttl, cfg.TTL)
	}
	if cfg.NoExpiration != false {
		t.Error("expected NoExpiration false when TTL is set")
	}
}

func TestWithInvalidation(t *testing.T) {
	opt := WithInvalidation()

	cfg := &SetConfig{}
	opt(cfg)

	if !cfg.PublishInvalidation {
		t.Error("expected PublishInvalidation true")
	}
}

func TestWithNoExpiration(t *testing.T) {
	opt := WithNoExpiration()

	cfg := &SetConfig{TTL: 10 * time.Second}
	opt(cfg)

	if cfg.TTL != 0 {
		t.Errorf("expected TTL 0, got %v", cfg.TTL)
	}
	if !cfg.NoExpiration {
		t.Error("expected NoExpiration true")
	}
}

func TestApplyOptionsNoOptions(t *testing.T) {
	cfg := ApplyOptions([]OptFnc{})

	if cfg.TTL != 0 {
		t.Errorf("expected TTL 0, got %v", cfg.TTL)
	}
	if cfg.PublishInvalidation != false {
		t.Error("expected PublishInvalidation false")
	}
	if cfg.NoExpiration != false {
		t.Error("expected NoExpiration false")
	}
}

func TestApplyOptionsSingleOption(t *testing.T) {
	cfg := ApplyOptions([]OptFnc{WithTTL(1 * time.Hour)})

	if cfg.TTL != 1*time.Hour {
		t.Errorf("expected TTL 1h, got %v", cfg.TTL)
	}
}

func TestApplyOptionsMultipleOptions(t *testing.T) {
	cfg := ApplyOptions([]OptFnc{
		WithTTL(30 * time.Minute),
		WithInvalidation(),
	})

	if cfg.TTL != 30*time.Minute {
		t.Errorf("expected TTL 30m, got %v", cfg.TTL)
	}
	if !cfg.PublishInvalidation {
		t.Error("expected PublishInvalidation true")
	}
}

func TestApplyOptionsOverride(t *testing.T) {
	cfg := ApplyOptions([]OptFnc{
		WithTTL(1 * time.Hour),
		WithTTL(2 * time.Hour),
	})

	if cfg.TTL != 2*time.Hour {
		t.Errorf("expected TTL 2h, got %v", cfg.TTL)
	}
}

func TestApplyOptionsNoExpirationOverridesTTL(t *testing.T) {
	cfg := ApplyOptions([]OptFnc{
		WithTTL(1 * time.Hour),
		WithNoExpiration(),
	})

	if cfg.TTL != 0 {
		t.Errorf("expected TTL 0, got %v", cfg.TTL)
	}
	if !cfg.NoExpiration {
		t.Error("expected NoExpiration true")
	}
}

func TestApplyOptionsTTLOverridesNoExpiration(t *testing.T) {
	cfg := ApplyOptions([]OptFnc{
		WithNoExpiration(),
		WithTTL(1 * time.Hour),
	})

	if cfg.TTL != 1*time.Hour {
		t.Errorf("expected TTL 1h, got %v", cfg.TTL)
	}
	if cfg.NoExpiration != false {
		t.Error("expected NoExpiration false when TTL is set after NoExpiration")
	}
}

func TestNoExpirationFlagBasic(t *testing.T) {
	opt := WithNoExpiration()

	cfg := &SetConfig{
		TTL:          5 * time.Minute,
		NoExpiration: false,
	}
	opt(cfg)

	if cfg.NoExpiration != true {
		t.Error("expected NoExpiration true")
	}
	if cfg.TTL != 0 {
		t.Errorf("expected TTL 0, got %v", cfg.TTL)
	}
}

func TestNoExpirationFlagWithMultipleOptions(t *testing.T) {
	cfg := ApplyOptions([]OptFnc{
		WithInvalidation(),
		WithNoExpiration(),
	})

	if !cfg.NoExpiration {
		t.Error("expected NoExpiration true")
	}
	if !cfg.PublishInvalidation {
		t.Error("expected PublishInvalidation true")
	}
	if cfg.TTL != 0 {
		t.Errorf("expected TTL 0, got %v", cfg.TTL)
	}
}

func TestWithTTLResetsNoExpiration(t *testing.T) {
	cfg := &SetConfig{
		NoExpiration: true,
		TTL:          0,
	}

	opt := WithTTL(10 * time.Minute)
	opt(cfg)

	if cfg.NoExpiration != false {
		t.Error("expected NoExpiration false after setting TTL")
	}
	if cfg.TTL != 10*time.Minute {
		t.Errorf("expected TTL 10m, got %v", cfg.TTL)
	}
}

func TestApplyOptionsOnlyNoExpiration(t *testing.T) {
	cfg := ApplyOptions([]OptFnc{
		WithNoExpiration(),
	})

	if !cfg.NoExpiration {
		t.Error("expected NoExpiration true")
	}
	if cfg.TTL != 0 {
		t.Errorf("expected TTL 0, got %v", cfg.TTL)
	}
	if cfg.PublishInvalidation {
		t.Error("expected PublishInvalidation false")
	}
}
