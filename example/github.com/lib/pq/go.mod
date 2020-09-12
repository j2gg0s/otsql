module github.com/j2gg0s/otsql/example/github.com/lib/pq

go 1.14

require (
	contrib.go.opencensus.io/integrations/ocsql v0.1.6
	github.com/j2gg0s/otsql v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.8.0
	go.opencensus.io v0.22.4 // indirect
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.11.0
	go.opentelemetry.io/otel/exporters/stdout v0.11.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.11.0
	go.opentelemetry.io/otel/sdk v0.11.0
)

replace github.com/j2gg0s/otsql => /Users/j2gg0s/go/src/github.com/j2gg0s/otsql
