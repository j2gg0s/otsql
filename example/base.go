package example

import (
	"net/http"

	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
)

func InitTracer() {
	_, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")),
	)

	if err != nil {
		panic(err)
	}
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
