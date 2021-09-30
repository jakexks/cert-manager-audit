// Package logrlogger is an audit logging sync that logs to the given logr.Logger
package logrlogger

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apiserver/pkg/apis/audit"

	"sandbox.jakexks.dev/cert-manager-audit/pkg/process"
	"sandbox.jakexks.dev/cert-manager-audit/pkg/sink"
)

type stdoutAdapter struct{}

func (*stdoutAdapter) New() sink.Sink {
	return &stdoutSink{}
}

func (*stdoutAdapter) Name() string {
	return "stdout"
}

type stdoutSink struct {
	logger logr.Logger
}

func (s *stdoutSink) Setup(logger logr.Logger, config sink.Config) (process.Func, error) {
	s.logger = logger.V(10)
	return func(ctx context.Context, events []*audit.Event) error {
		for _, e := range events {
			s.logger.Info("received audit event", "event", e)
		}
		return nil
	}, nil
}

func (*stdoutSink) Start(ctx context.Context) error {
	return nil
}

func (*stdoutSink) Stop(ctx context.Context) error {
	return nil
}

func init() {
	sink.Register(&stdoutAdapter{})
}
