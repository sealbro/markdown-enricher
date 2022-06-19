package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type PrometheusConfig struct {
	Port string
}

type PrometheusServer struct {
	config *PrometheusConfig
}

func MakePrometheusServer(config *PrometheusConfig) *PrometheusServer {
	return &PrometheusServer{
		config: config,
	}
}

func (s *PrometheusServer) Start() error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":"+s.config.Port, nil)
}
