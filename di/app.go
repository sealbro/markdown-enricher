package di

import (
	"context"
	"markdown-enricher/infrastructure/metrics"
	"markdown-enricher/infrastructure/trace"
	router "markdown-enricher/infrastructure/web"
	"markdown-enricher/pkg/closer"
	"markdown-enricher/pkg/graceful"
	"markdown-enricher/usecases/jobs"
)

type App struct {
	*graceful.Graceful
}

func MakeApplication(collection *closer.CloserCollection, server router.WebServer, prometheusServer *metrics.PrometheusServer, tp trace.TracerProvider, job *jobs.MdFileEnricherJob) graceful.Application {
	return &App{
		Graceful: &graceful.Graceful{
			StartAction: func() error {
				go prometheusServer.Start()
				go job.Start()

				return server.ListenAndServe()
			},
			DeferAction: func(ctx context.Context) error {
				return collection.Close(ctx)
			},
			ShutdownAction: func(ctx context.Context) error {
				return tp.Shutdown(ctx)
			},
		},
	}
}
