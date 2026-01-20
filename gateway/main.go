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

var tracer = otel.Tracer("gateway-service")

func initTracer() func() {
	ctx := context.Background()

	// Endpoint do Jaeger (OTLP HTTP)
	jaegerEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://jaeger:4318" 
	}

	log.Printf("🔧 Inicializando OpenTelemetry")
	log.Printf("   Endpoint: %s", jaegerEndpoint)
	log.Printf("   Service: gateway-service")

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
			semconv.ServiceName("gateway-service"),
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
	ctx := r.Context()

	// Cria um span
	ctx, span := tracer.Start(ctx, "gateway.getProducts")
	defer span.End()

	// Adiciona atributos ao span
	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
		attribute.String("component", "gateway"),
	)

	log.Printf("🔍 Processando requisição GET /products")

	// URL do serviço de produtos
	productsURL := os.Getenv("PRODUCTS_SERVICE_URL")
	if productsURL == "" {
		productsURL = "http://products:8081"
	}

	// Cria um child span para a chamada HTTP
	ctx, httpSpan := tracer.Start(ctx, "http.call.products-service")
	httpSpan.SetAttributes(
		attribute.String("http.url", productsURL+"/products"),
		attribute.String("peer.service", "products-service"),
	)

	// Cria requisição HTTP com contexto (propagação automática)
	req, err := http.NewRequestWithContext(ctx, "GET", productsURL+"/products", nil)
	if err != nil {
		httpSpan.RecordError(err)
		httpSpan.SetStatus(codes.Error, "failed to create request")
		httpSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Erro ao criar requisição", http.StatusInternalServerError)
		return
	}

	log.Printf("📡 Chamando products service em %s", productsURL+"/products")

	// Cliente HTTP com instrumentação automática (propagação de contexto)
	client := http.Client{
		Timeout:   5 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	resp, err := client.Do(req)
	httpSpan.End()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Erro ao chamar products service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Lê a resposta
	var products []Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Erro ao decodificar resposta", http.StatusInternalServerError)
		return
	}

	// Adiciona evento ao span
	span.SetAttributes(attribute.Int("products.count", len(products)))
	span.AddEvent("products processed")

	// Retorna os produtos
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)

	span.SetStatus(codes.Ok, "success")
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

	// Rotas (sem instrumentação automática para ter controle manual)
	mux.HandleFunc("/products", getProductsHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Wrap o handler com instrumentação OpenTelemetry
	handler := otelhttp.NewHandler(mux, "gateway-server")

	log.Println("🚀 Gateway Service rodando na porta 8080...")
	log.Println("📍 Endpoints disponíveis:")
	log.Println("   - GET  /products")
	log.Println("   - GET  /health")
	log.Println("   - GET  /metrics")
	log.Println("🔭 Tracing: OpenTelemetry → Jaeger")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
	}
}