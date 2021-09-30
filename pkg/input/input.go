// Package input contains the interfaces for audit logging input
package input

import (
	"context"
	"sync"

	"github.com/go-logr/logr"

	"sandbox.jakexks.dev/cert-manager-audit/pkg/process"
)

// All registered input adapters are kept here
var (
	inputs = make(map[string]Adapter)
	lock   sync.RWMutex
)

// An Adapter is an Input factory
type Adapter interface {
	New() Input
	Name() string
}

// The Input interface should be implemented by all adapters
type Input interface {
	// Setup should read any required config, check for dependencies, etc
	Setup(logr.Logger, process.Func, Config) error
	// Start should start the adapter listening for incoming events.
	// It must not block.
	Start(context.Context) error
	// Stop should stop accepting any incoming events and store as much as possible.
	Stop(context.Context) error
}

// Config is serialized configuration.
type Config []byte

func Register(adapter Adapter) {
	lock.Lock()
	defer lock.Unlock()
	inputs[adapter.Name()] = adapter
}

func Adapters() map[string]Adapter {
	lock.RLock()
	defer lock.RUnlock()
	out := make(map[string]Adapter)
	for k, v := range inputs {
		out[k] = v
	}
	return out
}
