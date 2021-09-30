package process

import (
	"context"

	"golang.org/x/sync/errgroup"
	"k8s.io/apiserver/pkg/apis/audit"
)

// Func is called by input adapters for each incoming event
type Func func(ctx context.Context, events []*audit.Event) error

// FanOut sends events to all known outputs
func FanOut(ctx context.Context, sinks []Func, events []*audit.Event) error {
	fanOut, fanOutCtx := errgroup.WithContext(ctx)
	for _, s := range sinks {
		fanOut.Go(func() error {
			return s(fanOutCtx, events)
		})
	}
	return fanOut.Wait()
}
