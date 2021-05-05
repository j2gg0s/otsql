package example

import (
	"net/http"

	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer() func() {
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcessFromEnv(),
		jaeger.WithSDKOptions(
			trace.WithSampler(trace.AlwaysSample()),
		),
	)

	if err != nil {
		panic(err)
	}
	return flush
}

func InitMeter() {
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{})
	if err != nil {
		panic(err)
	}

	go func() {
		err := http.ListenAndServe(":2222", exporter)
		if err != nil {
			panic(err)
		}
	}()
}
