package main

import (
	"context"
	"log"
	"time"
	"whapp-irc/config"
	"whapp-irc/whapp"
)

// A Bridge represents the bridging between an IRC connection and a WhatsApp web
// instance.
type Bridge struct {
	WI *whapp.Instance

	cfg     *config.Config
	started bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// MakeBridge makes and returns a new Bridge instance.
func MakeBridge(cfg *config.Config) *Bridge {
	return &Bridge{
		started: false,
		cfg:     cfg,
	}
}

// Start starts the current bridge instance.
func (b *Bridge) Start() (started bool, err error) {
	if b.started {
		return false, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	wi, err := whapp.MakeInstanceWithPool(ctx, pool, true, b.cfg)
	if err != nil {
		cancel()
		return false, err
	}

	b.started = true
	b.WI = wi
	b.ctx = ctx
	b.cancel = cancel

	return true, nil
}

// Stop stops the current bridge instance.
func (b *Bridge) Stop() (stopped bool) {
	if !b.started {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := b.WI.Shutdown(ctx); err != nil {
		// TODO: how do we handle this?
		log.Printf("error while shutting down: %s", err.Error())
	}

	b.cancel()
	cancel()

	b.cfg = nil
	b.started = false
	b.WI = nil
	b.ctx = nil
	b.cancel = nil

	return true
}
