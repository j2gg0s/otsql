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
	"github.com/j2gg0s/otsql/hook/metric"
	"github.com/j2gg0s/otsql/hook/trace"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	exporter "go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"

	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
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
	exporter, err := jaeger.NewRawExporter(
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
		oteltrace.WithBatcher(exporter),
		oteltrace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)
}

func InitMeter() {
	exporter, err := exporter.InstallNewPipeline(
		exporter.Config{
			Registry: prometheus.DefaultRegisterer.(*prometheus.Registry),
		},
		controller.WithResource(resource.Empty()))
	if err != nil {
		panic(err)
	}

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", prometheusPort), exporter)
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
		),
	)
	if err != nil {
		return "", fmt.Errorf("register driver: %w", err)
	}

	return newName, nil
}
