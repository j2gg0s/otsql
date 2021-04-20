package otsql

import (
	"context"
	"database/sql/driver"
	"io/ioutil"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/trace"
)

var (
	query = "SELECT * FROM user WHERE name = :name LIMIT :limit"
	args  interface{}
)

func benchTrace(b *testing.B, ctx context.Context, options ...TraceOption) {
	o := newTraceOptions(options...)
	for i := 0; i < b.N; i++ {
		afterCtx, _, endTrace := startTrace(ctx, o, methodQuery, query, args)
		endTrace(afterCtx, nil)
	}
}

func withParent(b *testing.B) {
	ctx, _ := otel.GetTracerProvider().
		Tracer("github.com/j2gg0s/otsql").
		Start(context.Background(), "root", trace.WithNewRoot())
	benchTrace(b, ctx)
}

func newRoot(b *testing.B) {
	benchTrace(b, context.Background(), WithAllowRoot(true))
}

func withQuery(b *testing.B) {
	ctx, _ := otel.GetTracerProvider().
		Tracer("github.com/j2gg0s/otsql").
		Start(context.Background(), "root", trace.WithNewRoot())
	benchTrace(b, ctx, WithQuery(true))
}

func withValue(b *testing.B) {
	ctx, _ := otel.GetTracerProvider().
		Tracer("github.com/j2gg0s/otsql").
		Start(context.Background(), "root", trace.WithNewRoot())
	benchTrace(b, ctx, WithQuery(true), WithQueryParams(true))
}

func withDefaultLabels(b *testing.B) {
	ctx, _ := otel.GetTracerProvider().
		Tracer("github.com/j2gg0s/otsql").
		Start(context.Background(), "root", trace.WithNewRoot())
	benchTrace(b, ctx, WithDefaultAttributes(attribute.String("A", "a"), attribute.String("B", "b")))
}

func BenchmarkTrace(b *testing.B) {
	b.Run("Without parent and exporter", newRoot)

	stdout.InstallNewPipeline([]stdout.Option{stdout.WithWriter(ioutil.Discard)}, nil)
	args = []driver.NamedValue{
		driver.NamedValue{
			Name:    "",
			Ordinal: 0,
			Value:   "j2gg0s",
		},
		driver.NamedValue{
			Name:    "",
			Ordinal: 1,
			Value:   10,
		},
	}
	b.Run("Without parent span.", newRoot)
	b.Run("With parent span", withParent)
	b.Run("With query", withQuery)
	b.Run("With NamedValue args", withValue)
	b.Run("With default attributes", withDefaultLabels)

	args = []driver.Value{
		"j2gg0s",
		10,
	}
	b.Run("With Value args", withValue)
}
