package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

func initJaeger(serviceName string) (opentracing.Tracer, io.Closer) {
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: fmt.Sprintf("%s:%s", os.Getenv("JAEGER_AGENT_HOST"), os.Getenv("JAEGER_AGENT_PORT")),
		},
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		log.Fatalf("Erro ao inicializar Jaeger: %v", err)
	}

	opentracing.SetGlobalTracer(tracer)
	return tracer, closer
}

func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("gateway-get-products")
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	// Simula processamento no gateway
	span.LogKV("event", "validating request")
	time.Sleep(50 * time.Millisecond)

	// Chama o serviço de produtos
	productsURL := os.Getenv("PRODUCTS_SERVICE_URL") + "/products"
	
	req, err := http.NewRequestWithContext(ctx, "GET", productsURL, nil)
	if err != nil {
		span.SetTag("error", true)
		span.LogKV("error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Injeta o contexto do trace na requisição HTTP
	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, productsURL)
	ext.HTTPMethod.Set(span, "GET")
	
	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	err = tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	if err != nil {
		log.Printf("Erro ao injetar contexto: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		span.SetTag("error", true)
		span.LogKV("error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.SetTag("error", true)
		span.LogKV("error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.LogKV("event", "products retrieved successfully")
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	_, closer := initJaeger("gateway-service")
	defer closer.Close()

	http.HandleFunc("/api/products", getProductsHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("Gateway rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}