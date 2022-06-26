package logger

import (
	"github.com/labstack/echo/v4/middleware"
)

var EchoLoggerConfig = middleware.LoggerConfig{
	Skipper: middleware.DefaultSkipper,
	Format: `{"ts_orig":"${time_rfc3339}", "level":"DEBUG", "trace_id":"${header:trace-id}", "remote_ip":"${remote_ip}",` +
		`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
		`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
		`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
	CustomTimeFormat: "2006-01-02 15:04:05.00000",
}
