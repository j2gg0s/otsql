package example

import (
	"net/http"

	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"

	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
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
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{}, controller.WithResource(resource.Empty()))
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
