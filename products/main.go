package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var products = []Product{
	{ID: 1, Name: "Notebook", Price: 3500.00},
	{ID: 2, Name: "Mouse", Price: 50.00},
	{ID: 3, Name: "Teclado", Price: 150.00},
	{ID: 4, Name: "Monitor", Price: 1200.00},
}

var tracer = otel.Tracer("products-service")

func initTracer() func() {
	ctx := context.Background()

	// Endpoint do Jaeger (OTLP HTTP)
	jaegerEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://jaeger:4318" // Porta OTLP HTTP
	}

	log.Printf("🔧 Inicializando OpenTelemetry")
	log.Printf("   Endpoint: %s", jaegerEndpoint)
	log.Printf("   Service: products-service")

	// Cria o exporter OTLP HTTP
	exporter, err := otlptrace.New(
		ctx,
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(jaegerEndpoint[7:]), 
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		log.Fatalf("❌ Erro ao criar OTLP exporter: %v", err)
	}

	// Cria o resource com informações do serviço
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("products-service"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "development"),
		),
	)
	if err != nil {
		log.Fatalf("❌ Erro ao criar resource: %v", err)
	}

	// Cria o TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), 
	)

	// Registra o TracerProvider global
	otel.SetTracerProvider(tp)

	// Configura propagação de contexto (W3C Trace Context)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	log.Printf("✅ OpenTelemetry inicializado com sucesso!")

	// Retorna função de cleanup
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("⚠️  Erro ao fazer shutdown do TracerProvider: %v", err)
		}
	}
}

func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	// O contexto já vem com o trace propagado pelo middleware otelhttp
	ctx := r.Context()

	// Cria um span (vai ser child do span do gateway se propagado)
	ctx, span := tracer.Start(ctx, "products.getAll")
	defer span.End()

	// Adiciona atributos ao span
	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
		attribute.String("component", "products"),
	)

	log.Printf("🔍 Processando requisição GET /products")

	// Simula busca no banco de dados
	ctx, dbSpan := tracer.Start(ctx, "database.query")
	dbSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", "SELECT * FROM products"),
		attribute.String("db.operation", "SELECT"),
	)
	span.AddEvent("fetching products from database")
	time.Sleep(100 * time.Millisecond)
	dbSpan.End()

	// Simula processamento/filtro
	ctx, filterSpan := tracer.Start(ctx, "products.filter")
	filterSpan.SetAttributes(
		attribute.String("filter.type", "price-range"),
	)
	span.AddEvent("filtering products")
	time.Sleep(30 * time.Millisecond)
	filterSpan.End()

	// Adiciona métricas ao span
	span.SetAttributes(attribute.Int("products.count", len(products)))
	span.AddEvent("products processed")

	// Retorna os produtos
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Erro ao encodar resposta", http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "success")
	span.AddEvent("response sent successfully")
	log.Printf("✅ Requisição completada - %d produtos retornados", len(products))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	// Inicializa OpenTelemetry
	shutdown := initTracer()
	defer shutdown()

	// Cria mux com instrumentação automática
	mux := http.NewServeMux()

	// Rotas
	mux.HandleFunc("/products", getProductsHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Wrap o handler com instrumentação OpenTelemetry (propagação automática)
	handler := otelhttp.NewHandler(mux, "products-server")

	log.Println("🚀 Products Service rodando na porta 8081...")
	log.Println("📍 Endpoints disponíveis:")
	log.Println("   - GET  /products")
	log.Println("   - GET  /health")
	log.Println("   - GET  /metrics")
	log.Println("🔭 Tracing: OpenTelemetry → Jaeger")

	if err := http.ListenAndServe(":8081", handler); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
	}
}