package main

import (
	"markdown-enricher/di"
	_ "markdown-enricher/docs"
	"markdown-enricher/pkg/graceful"
)

// @title Markdown enricher
// @version 1.0
// @host localhost:8080
// @BasePath /api
// @schemes http
func main() {
	container := di.Build()

	err := container.Invoke(func(application graceful.Application) {
		application.RunAndWait()
	})

	if err != nil {
		panic(err)
	}
}
