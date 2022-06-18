package di

import (
	"context"
	"go.uber.org/dig"
	"log"
	"markdown-enricher/infrastructure/cache"
	"markdown-enricher/infrastructure/db"
	router "markdown-enricher/infrastructure/web"
	"markdown-enricher/pkg/closer"
	"markdown-enricher/pkg/env"
	"markdown-enricher/pkg/graceful"
	controller "markdown-enricher/usecases/controllers"
	"markdown-enricher/usecases/interactors"
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

	provideOrPanic(container, parsers.MakeMarkdownParser)
	provideOrPanic(container, services.MakeGitHubService)
	provideOrPanic(container, cache.MakeMemoryCacheService)
	// DB
	provideOrPanic(container, func() *db.SqliteConfig {
		return &db.SqliteConfig{Connection: env.EnvOrDefault("SQLITE_CONNECTION", "markdown.db")}
	})
	provideOrPanic(container, db.MakeSqliteConnection)
	// UseCases
	provideOrPanic(container, repositories.MakeSqlStorageRepository)
	provideOrPanic(container, interactors.MakeEnricherInteractor)
	provideOrPanic(container, closer.MakeCloserCollection)
	// API
	provideOrPanic(container, controller.MakeMarkdownController, dig.Group("controller"))
	provideOrPanic(container, func(group ControllerGroup) []router.Controller { return group.Controllers })
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

type App struct {
	*graceful.Graceful
}

func MakeApplication(collection *closer.CloserCollection, server router.WebServer) graceful.Application {
	return &App{
		Graceful: &graceful.Graceful{
			StartAction: func() error {
				return server.ListenAndServe()
			},
			DeferAction: func(ctx context.Context) error {
				return collection.Close(ctx)
			},
			ShutdownAction: func(ctx context.Context) error {
				return nil
			},
		},
	}
}
