package example

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/j2gg0s/otsql"
	"github.com/j2gg0s/otsql/hook/log"
	"github.com/j2gg0s/otsql/hook/metric"
	"github.com/j2gg0s/otsql/hook/trace"
	"github.com/prometheus/client_golang/prometheus"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	exporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	export "go.opentelemetry.io/otel/sdk/export/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var (
	jaegerCollectorEndpoint = "http://localhost:14268/api/traces"
	jaegerAPIEndpoint       = "http://localhost:16686/api/traces"
	serviceName             = "otsql-example"
	prometheusPort          = "2222"

	MySQLDSN      = "otsql_user:otsql_password@/otsql_db?parseTime=true"
	PostgreSQLDSN = "user=otsql_user password=otsql_password dbname=otsql_db host=localhost port=5432 sslmode=disable TimeZone=Asia/Shanghai"
)

func InitTracer() {
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerCollectorEndpoint)),
	)
	if err != nil {
		panic(err)
	}

	resource, err := resource.New(
		context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
	)
	if err != nil {
		panic(err)
	}

	tp := oteltrace.NewTracerProvider(
		oteltrace.WithBatcher(exp),
		oteltrace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)
}

func InitMeter() {
	config := exporter.Config{}
	ctl := controller.New(
		processor.New(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			export.CumulativeExportKindSelector(),
			processor.WithMemory(true),
		),
		controller.WithResource(resource.Empty()),
	)
	exp, err := exporter.New(
		exporter.Config{Registry: prometheus.DefaultRegisterer.(*prometheus.Registry)},
		ctl,
	)
	if err != nil {
		panic(err)
	}
	global.SetMeterProvider(exp.MeterProvider())

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", prometheusPort), exp)
		if err != nil {
			panic(err)
		}
	}()
}

func PrintTraces(ctx context.Context, httpClient *http.Client) {
	vals := url.Values{}
	vals.Set("limit", "20")
	vals.Set("lookback", "1m")
	vals.Set("service", serviceName)
	addr := fmt.Sprintf("%s?%s", jaegerAPIEndpoint, vals.Encode())

	resp, err := httpClient.Get(addr)
	if err != nil {
		fmt.Printf("GET %s: %s\n", addr, err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GET %s: Status(%s)\n", addr, resp.Status)
		return
	}
	fmt.Printf("GET %s:\n", addr)
	m := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		fmt.Printf("unmarshal: %s", err.Error())
		return
	}
	body, err := json.MarshalIndent(&m, "", "  ")
	if err != nil {
		fmt.Printf("read body: %s\n", err.Error())
		return
	}
	fmt.Println(string(body))
}

func PrintMetrics(ctx context.Context, httpClient *http.Client) {
	addr := fmt.Sprintf("http://localhost:%s/metrics", prometheusPort)
	resp, err := httpClient.Get(addr)
	if err != nil {
		fmt.Printf("GET %s: %s\n", addr, err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GET %s: Status(%s)\n", addr, resp.Status)
		return
	}
	fmt.Printf("GET %s:\n", addr)
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "go_sql") {
			fmt.Println(line)
		}
	}
}

func AsyncCronPrint(every time.Duration) {
	go func() {
		ticker := time.NewTicker(every)
		for range ticker.C {
			PrintTraces(context.Background(), http.DefaultClient)
			PrintMetrics(context.Background(), http.DefaultClient)
		}
	}()
}

func Register(name string) (string, error) {
	metricHook, err := metric.New()
	if err != nil {
		return "", fmt.Errorf("new metric hook: %w", err)
	}

	newName, err := otsql.Register(
		name,
		otsql.WithHooks(
			trace.New(
				trace.WithAllowRoot(true),
				trace.WithQuery(true),
				trace.WithQueryParams(true),
			),
			metricHook,
			log.New(),
		),
	)
	if err != nil {
		return "", fmt.Errorf("register driver: %w", err)
	}

	return newName, nil
}
