package metrics

// https://github.com/labstack/echo-contrib/blob/master/prometheus/prometheus.go

import (
	"bytes"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"markdown-enricher/pkg/logger"
	"net/http"
	"strconv"
	"time"
)

var defaultSubsystem = "echo"

const (
	_          = iota // ignore first value by assigning to blank identifier
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
)

// reqDurBuckets is the buckets for request duration. Here, we use the prometheus defaults
// which are for ~10s request length max: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
var reqDurBuckets = prometheus.DefBuckets

// reqSzBuckets is the buckets for request size. Here we define a spectrom from 1KB thru 1NB up to 10MB.
var reqSzBuckets = []float64{1.0 * KB, 2.0 * KB, 5.0 * KB, 10.0 * KB, 100 * KB, 500 * KB, 1.0 * MB, 2.5 * MB, 5.0 * MB, 10.0 * MB}

// resSzBuckets is the buckets for response size. Here we define a spectrom from 1KB thru 1NB up to 10MB.
var resSzBuckets = []float64{1.0 * KB, 2.0 * KB, 5.0 * KB, 10.0 * KB, 100 * KB, 500 * KB, 1.0 * MB, 2.5 * MB, 5.0 * MB, 10.0 * MB}

// Standard default metrics
//	counter, counter_vec, gauge, gauge_vec,
//	histogram, histogram_vec, summary, summary_vec
var reqCnt = &Metric{
	ID:          "reqCnt",
	Name:        "requests_total",
	Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
	Type:        "counter_vec",
	Args:        []string{"code", "method", "host", "url"}}

var reqDur = &Metric{
	ID:          "reqDur",
	Name:        "request_duration_seconds",
	Description: "The HTTP request latencies in seconds.",
	Args:        []string{"code", "method", "url"},
	Type:        "histogram_vec",
	Buckets:     reqDurBuckets}

var resSz = &Metric{
	ID:          "resSz",
	Name:        "response_size_bytes",
	Description: "The HTTP response sizes in bytes.",
	Args:        []string{"code", "method", "url"},
	Type:        "histogram_vec",
	Buckets:     resSzBuckets}

var reqSz = &Metric{
	ID:          "reqSz",
	Name:        "request_size_bytes",
	Description: "The HTTP request sizes in bytes.",
	Args:        []string{"code", "method", "url"},
	Type:        "histogram_vec",
	Buckets:     reqSzBuckets}

var standardMetrics = []*Metric{
	reqCnt,
	reqDur,
	resSz,
	reqSz,
}

/*
RequestCounterLabelMappingFunc is a function which can be supplied to the middleware to control
the cardinality of the request counter's "url" label, which might be required in some contexts.
For instance, if for a "/customer/:name" route you don't want to generate a time series for every
possible customer name, you could use this function:

func(c echo.Context) string {
	url := c.Request.URL.Path
	for _, p := range c.Params {
		if p.Key == "name" {
			url = strings.Replace(url, p.Value, ":name", 1)
			break
		}
	}
	return url
}

which would map "/customer/alice" and "/customer/bob" to their template "/customer/:name".
It can also be applied for the "Host" label
*/
type RequestCounterLabelMappingFunc func(c echo.Context) string

// Metric is a definition for the name, description, type, ID, and
// prometheus.Collector type (i.e. CounterVec, Summary, etc) of each metric
type Metric struct {
	MetricCollector prometheus.Collector
	ID              string
	Name            string
	Description     string
	Type            string
	Args            []string
	Buckets         []float64
}

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqCnt               *prometheus.CounterVec
	reqDur, reqSz, resSz *prometheus.HistogramVec
	router               *echo.Echo
	listenAddress        string
	Ppg                  PushGateway

	MetricsList []*Metric
	Subsystem   string
	Skipper     middleware.Skipper

	RequestCounterURLLabelMappingFunc  RequestCounterLabelMappingFunc
	RequestCounterHostLabelMappingFunc RequestCounterLabelMappingFunc

	// Context string to use as a prometheus URL label
	URLLabelFromContext string
}

// PushGateway contains the configuration for pushing to a Prometheus pushgateway (optional)
type PushGateway struct {
	// Push interval in seconds
	//lint:ignore ST1011 renaming would be breaking change
	PushIntervalSeconds time.Duration

	// Push Gateway URL in format http://domain:port
	// where JOBNAME can be any string of your choice
	PushGatewayURL string

	// pushgateway job name, defaults to "echo"
	Job string
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus(subsystem string, skipper middleware.Skipper, customMetricsList ...[]*Metric) *Prometheus {
	var metricsList []*Metric
	if skipper == nil {
		skipper = middleware.DefaultSkipper
	}

	if len(customMetricsList) > 1 {
		panic("Too many args. NewPrometheus( string, <optional []*Metric> ).")
	} else if len(customMetricsList) == 1 {
		metricsList = customMetricsList[0]
	}

	metricsList = append(metricsList, standardMetrics...)

	p := &Prometheus{
		MetricsList: metricsList,
		Subsystem:   defaultSubsystem,
		Skipper:     skipper,
		RequestCounterURLLabelMappingFunc: func(c echo.Context) string {
			return c.Path() // i.e. by default do nothing, i.e. return URL as is
		},
		RequestCounterHostLabelMappingFunc: func(c echo.Context) string {
			return c.Request().Host
		},
	}

	p.registerMetrics(subsystem)

	return p
}

func (p *Prometheus) runServer() {
	if p.listenAddress != "" {
		go p.router.Start(p.listenAddress)
	}
}

func (p *Prometheus) getMetrics() []byte {
	out := &bytes.Buffer{}
	metricFamilies, _ := prometheus.DefaultGatherer.Gather()
	for i := range metricFamilies {
		expfmt.MetricFamilyToText(out, metricFamilies[i])

	}
	return out.Bytes()
}

// NewMetric associates prometheus.Collector based on Metric.Type
func NewMetric(m *Metric, subsystem string) prometheus.Collector {
	var metric prometheus.Collector
	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "counter":
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "gauge_vec":
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "gauge":
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "histogram_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
				Buckets:   m.Buckets,
			},
			m.Args,
		)
	case "histogram":
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
				Buckets:   m.Buckets,
			},
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "summary":
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}
	return metric
}

func (p *Prometheus) registerMetrics(subsystem string) {

	for _, metricDef := range p.MetricsList {
		metric := NewMetric(metricDef, subsystem)
		if err := prometheus.Register(metric); err != nil {
			logger.Errorf("%s could not be registered in Prometheus: %v", metricDef.Name, err)
		}
		switch metricDef {
		case reqCnt:
			p.reqCnt = metric.(*prometheus.CounterVec)
		case reqDur:
			p.reqDur = metric.(*prometheus.HistogramVec)
		case resSz:
			p.resSz = metric.(*prometheus.HistogramVec)
		case reqSz:
			p.reqSz = metric.(*prometheus.HistogramVec)
		}
		metricDef.MetricCollector = metric
	}
}

// Use adds the middleware to the Echo engine.
func (p *Prometheus) Use(e *echo.Echo) {
	e.Use(p.HandlerFunc)
}

// HandlerFunc defines handler function for middleware
func (p *Prometheus) HandlerFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if p.Skipper(c) {
			return next(c)
		}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request())

		err := next(c)

		status := c.Response().Status
		if err != nil {
			var httpError *echo.HTTPError
			if errors.As(err, &httpError) {
				status = httpError.Code
			}
			if status == 0 || status == http.StatusOK {
				status = http.StatusInternalServerError
			}
		}

		elapsed := float64(time.Since(start)) / float64(time.Second)

		url := p.RequestCounterURLLabelMappingFunc(c)
		if len(p.URLLabelFromContext) > 0 {
			u := c.Get(p.URLLabelFromContext)
			if u == nil {
				u = "unknown"
			}
			url = u.(string)
		}

		statusStr := strconv.Itoa(status)
		p.reqDur.WithLabelValues(statusStr, c.Request().Method, url).Observe(elapsed)
		p.reqCnt.WithLabelValues(statusStr, c.Request().Method, p.RequestCounterHostLabelMappingFunc(c), url).Inc()
		p.reqSz.WithLabelValues(statusStr, c.Request().Method, url).Observe(float64(reqSz))

		resSz := float64(c.Response().Size)
		p.resSz.WithLabelValues(statusStr, c.Request().Method, url).Observe(resSz)

		return err
	}
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
