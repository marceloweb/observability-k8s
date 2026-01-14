package main

import (
	"encoding/json"
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
	
	// Extrai o contexto do trace da requisição HTTP
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	
	span := tracer.StartSpan("products-get-all", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	ext.HTTPMethod.Set(span, r.Method)
	ext.HTTPUrl.Set(span, r.URL.String())
	
	span.LogKV("event", "fetching products from database")
	
	// Simula acesso ao banco de dados
	time.Sleep(100 * time.Millisecond)
	
	// Simula filtro/processamento
	span.LogKV("event", "filtering products")
	time.Sleep(30 * time.Millisecond)
	
	span.SetTag("products.count", len(products))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
	
	ext.HTTPStatusCode.Set(span, 200)
	span.LogKV("event", "response sent successfully")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	_, closer := initJaeger("products-service")
	defer closer.Close()

	http.HandleFunc("/products", getProductsHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("Products Service rodando na porta 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}