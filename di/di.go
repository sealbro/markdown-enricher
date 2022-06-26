package di

import (
	"go.uber.org/dig"
	"log"
	"markdown-enricher/infrastructure/cache"
	"markdown-enricher/infrastructure/db"
	"markdown-enricher/infrastructure/metrics"
	"markdown-enricher/infrastructure/trace"
	router "markdown-enricher/infrastructure/web"
	"markdown-enricher/pkg/closer"
	"markdown-enricher/pkg/env"
	controller "markdown-enricher/usecases/controllers"
	"markdown-enricher/usecases/interactors"
	"markdown-enricher/usecases/jobs"
	"markdown-enricher/usecases/parsers"
	"markdown-enricher/usecases/repositories"
	"markdown-enricher/usecases/services"
)

type ControllerGroup struct {
	dig.In
	Controllers []router.Controller `group:"controller"`
}

func Build() *dig.Container {
	container := dig.New()

	provideOrPanic(container, closer.MakeCloserCollection)
	provideOrPanic(container, cache.MakeMemoryCacheService)
	// DB
	provideOrPanic(container, func() *db.SqliteConfig {
		return &db.SqliteConfig{Connection: env.EnvOrDefault("SQLITE_CONNECTION", "markdown.db")}
	})
	provideOrPanic(container, db.MakeSqliteConnection)
	// UseCases
	provideOrPanic(container, parsers.MakeMarkdownParser)
	provideOrPanic(container, services.MakeGitHubService)
	provideOrPanic(container, interactors.MakeFilesToProcessQueue)
	provideOrPanic(container, repositories.MakeSqlLinkRepository)
	provideOrPanic(container, repositories.MakeSqlMdFileRepository)
	provideOrPanic(container, interactors.MakeEnricherInteractor)
	provideOrPanic(container, jobs.MakeMdFileEnricherJob)
	// API
	provideOrPanic(container, controller.MakeMarkdownController, dig.Group("controller"))
	provideOrPanic(container, func(group ControllerGroup) []router.Controller { return group.Controllers })
	provideOrPanic(container, func() *trace.JaegerConfig {
		// "http://localhost:14268/api/traces"
		return &trace.JaegerConfig{AgentHost: env.EnvOrDefault("OTEL_EXPORTER_JAEGER_AGENT_HOST", "")}
	})
	provideOrPanic(container, trace.MakeTraceProvider)
	provideOrPanic(container, func() *metrics.PrometheusConfig {
		return &metrics.PrometheusConfig{Port: env.EnvOrDefault("METRICS_PORT", "2112")}
	})
	provideOrPanic(container, metrics.MakePrometheusServer)
	provideOrPanic(container, func() *router.WebServerConfig {
		return &router.WebServerConfig{Port: env.EnvOrDefault("PORT", "8080"), PathPrefix: "/api"}
	})
	provideOrPanic(container, router.MakeEchoWebServer)
	provideOrPanic(container, MakeApplication)

	return container
}

func provideOrPanic(container *dig.Container, constructor interface{}, opts ...dig.ProvideOption) {
	err := container.Provide(constructor, opts...)
	if err != nil {
		log.Fatalf("container.Provide: %v", err)
	}
}
