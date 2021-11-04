package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	jaegerEndpoint = "http://localhost:14268/api/traces"
	serviceName    = "fibonacci"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func parseArguments() (int, error) {
	args := os.Args[1:]
	if len(args) == 0 {
		return 0, fmt.Errorf("expected an int argument")
	}

	n, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("can't parse argument as integer: %w", err)
	}

	return n, nil
}

func createAndRegisterExporters() error {
	stdExporter, err := stdout.NewExporter(
		stdout.WithPrettyPrint(),
	)

	if err != nil {
		return err
	}

	jaegerExporter, err := jaeger.NewRawExporter(jaeger.WithCollectorEndpoint(jaegerEndpoint), jaeger.WithProcess(jaeger.Process{ServiceName: serviceName}))
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(stdExporter),
		sdktrace.WithSyncer(jaegerExporter))
	otel.SetTracerProvider(tp)

	return nil
}

func main() {
	err := createAndRegisterExporters()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Brows to localhost:3000_n=6")

	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(fibHandler), "root"))

	log.Fatal(http.ListenAndServe(":3000", nil))
}

func fibHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	var n int

	if len(req.URL.Query()["n"]) != 1 {
		err = fmt.Errorf("wrong number of orguments")
	} else {
		n, err = strconv.Atoi(req.URL.Query()["n"][0])
	}

	if err != nil {
		http.Error(w, "could not parse index n", 400)
		return
	}

	ctx := req.Context()
	result := <-Fibonacci(ctx, n)

	sp := trace.SpanFromContext(ctx)
	sp.SetAttributes(label.Int("parameter", n), label.Int("result", result))

	fmt.Fprintln(w, result)
}

func Fibonacci(ctx context.Context, n int) chan int {
	ch := make(chan int)

	go func() {
		tr := otel.GetTracerProvider().Tracer(serviceName)

		cctx, sp := tr.Start(ctx, fmt.Sprintf("Fibonacci(%d", n), trace.WithAttributes(label.Int("n", n)))
		defer sp.End()

		result := 1
		if n > 1 {
			a := Fibonacci(cctx, n-1)
			b := Fibonacci(cctx, n-2)
			result = <-a + <-b
		}

		sp.SetAttributes(label.Int("result", result))

		ch <- result
	}()

	return ch
}
