package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-logr/stdr"
	"golang.org/x/sync/errgroup"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/apiserver/pkg/server"

	"sandbox.jakexks.dev/cert-manager-audit/pkg/input"
	"sandbox.jakexks.dev/cert-manager-audit/pkg/process"
	"sandbox.jakexks.dev/cert-manager-audit/pkg/sink"

	// import any adapters you want to use
	_ "sandbox.jakexks.dev/cert-manager-audit/pkg/input/audit"
	_ "sandbox.jakexks.dev/cert-manager-audit/pkg/sink/logrlogger"
)

func main() {
	stdr.SetVerbosity(10)
	log := stdr.New(log.Default())

	ctx := server.SetupSignalContext()

	var activeInputs []input.Input
	var activeSinks []sink.Sink
	var processFuncs []process.Func

	for name, sinkAdapter := range sink.Adapters() {
		sink := sinkAdapter.New()
		processFunc, err := sink.Setup(log.WithName(name), []byte{})
		if err != nil {
			log.Error(err, "Could not Setup sink", "adapter", name)
			os.Exit(1)
		}
		err = sink.Start(ctx)
		if err != nil {
			log.Error(err, "Could not Start sink", "adapter", name)
			os.Exit(1)
		}
		activeSinks = append(activeSinks, sink)
		processFuncs = append(processFuncs, processFunc)
	}

	for name, inputAdapter := range input.Adapters() {
		input := inputAdapter.New()
		err := input.Setup(log.WithName(name), func(ctx context.Context, events []*audit.Event) error {
			return process.FanOut(ctx, processFuncs, events)
		}, []byte{})
		if err != nil {
			log.Error(err, "Could not Setup input", "adapter", name)
			os.Exit(1)
		}
		activeInputs = append(activeInputs, input)
		err = input.Start(ctx)
		if err != nil {
			log.Error(err, "Could not Start input", "adapter", name)
			os.Exit(1)
		}
	}

	<-ctx.Done()
	log.Info("Shutting down")
	quitContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	shutdownGroup, shutdownCtx := errgroup.WithContext(quitContext)
	for _, i := range activeInputs {
		shutdownGroup.Go(func() error {
			return i.Stop(shutdownCtx)
		})
	}
	for _, s := range activeSinks {
		shutdownGroup.Go(func() error {
			return s.Stop(shutdownCtx)
		})
	}
	shutdownGroup.Wait()
	log.Info("Shutdown complete")
}
