// Package sink contains the interfaces for audit logging output
package sink

import (
	"context"
	"sync"

	"github.com/go-logr/logr"

	"sandbox.jakexks.dev/cert-manager-audit/pkg/process"
)

// All registered sink adapters are kept here
var (
	sinks = make(map[string]Adapter)
	lock  sync.RWMutex
)

// An Adapter is a Sink factory
type Adapter interface {
	New() Sink
	Name() string
}

// Config is serialized configuration.
type Config []byte

// The Sink interface should be implemented by all output adapters
type Sink interface {
	// Setup should read any required config, check for dependencies, etc
	Setup(logr.Logger, Config) (process.Func, error)
	// Start should start the adapter listening for incoming events.
	// It must not block.
	Start(context.Context) error
	// Stop should stop accepting any incoming events and store as much as possible.
	Stop(context.Context) error
}

func Register(adapter Adapter) {
	lock.Lock()
	defer lock.Unlock()
	sinks[adapter.Name()] = adapter
}

func Adapters() map[string]Adapter {
	lock.RLock()
	defer lock.RUnlock()
	out := make(map[string]Adapter)
	for k, v := range sinks {
		out[k] = v
	}
	return out
}
