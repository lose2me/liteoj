package config

import "sync/atomic"

// RuntimeConfig is a thin atomic holder so the running server can swap config
// in place after /admin/settings saves config.toml.
type RuntimeConfig struct {
	v atomic.Value
}

func NewRuntimeConfig(cfg *Config) *RuntimeConfig {
	rc := &RuntimeConfig{}
	rc.Store(cfg)
	return rc
}

func (r *RuntimeConfig) Current() *Config {
	if r == nil {
		return nil
	}
	if cfg, ok := r.v.Load().(*Config); ok {
		return cfg
	}
	return nil
}

func (r *RuntimeConfig) Store(cfg *Config) {
	if r == nil || cfg == nil {
		return
	}
	r.v.Store(cfg)
}
