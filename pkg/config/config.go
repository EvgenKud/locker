package config

import "time"

type option func(*cfg)

func WithTTL(t time.Duration) option {
	return func(c *cfg) {
		c.ttl = t
	}
}

func (c *cfg) TTL() time.Duration {
	return c.ttl
}

func New(options ...option) *cfg {
	c := new(cfg)
	for _, o := range options {
		o(c)
	}
	return c
}
