package web

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/trace"
	"markdown-enricher/infrastructure/metrics"
	"markdown-enricher/pkg/logger"
	"path"
)

type WebServer interface {
	ListenAndServe() error
}

type Controller interface {
	RegisterRoutes(e *echo.Group)
}

type Route struct {
	Method  string
	Path    string
	Handler echo.HandlerFunc
}

type WebServerConfig struct {
	Port       string
	PathPrefix string
}

type EchoWebServer struct {
	echo      *echo.Echo
	port      string
	apiPrefix string
}

func MakeEchoWebServer(config *WebServerConfig, controllers []Controller) WebServer {
	server := &EchoWebServer{
		echo:      echo.New(),
		port:      config.Port,
		apiPrefix: path.Join(config.PathPrefix, "/api"),
	}

	// Metrics
	p := metrics.NewPrometheus("echo", nil)
	p.Use(server.echo)

	// Swagger
	server.echo.GET("/swagger/*", echoSwagger.WrapHandler)

	// Other middlewares
	serviceName := "markdown-enricher"
	server.echo.Use(otelecho.Middleware(serviceName))
	server.echo.Use(tracingEnrichMiddleware())
	server.echo.Use(middleware.LoggerWithConfig(logger.EchoLoggerConfig))
	server.echo.Use(middleware.Recover())

	// Register routes
	for _, c := range controllers {
		c.RegisterRoutes(server.echo.Group(server.apiPrefix))
	}

	return server
}

func tracingEnrichMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			const headerName = "trace-id"

			request := c.Request()
			externalTraceId := request.Header.Get(headerName)
			if externalTraceId == "" {
				ctx := request.Context()
				traceID := trace.SpanFromContext(ctx).SpanContext().TraceID()
				if traceID.IsValid() {
					request.Header.Set(headerName, traceID.String())
					c.SetRequest(request)
				}
			}

			return next(c)
		}
	}
}

func (server *EchoWebServer) ListenAndServe() error {
	return server.echo.Start(":" + server.port)
}
