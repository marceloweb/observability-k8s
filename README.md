# 🔭 Exemplo de Observabilidade com OpenTelemetry + Jaeger

Este projeto demonstra uma **stack completa de observabilidade** usando **OpenTelemetry** (padrão CNCF) para rastreamento distribuído, métricas e logs entre microsserviços.

## 🏗️ Arquitetura

### **Microsserviços**
- **Gateway Service** (porta 8080): Serviço que recebe requisições externas e orquestra chamadas
- **Products Service** (porta 8081): Serviço que gerencia produtos (API interna)

### **Stack de Observabilidade**

#### 📊 **Traces (Rastreamento Distribuído)**
- **Jaeger All-in-One** (UI: 16686, OTLP: 4318): Plataforma completa para distributed tracing
  - **Collector**: Recebe traces via OTLP (OpenTelemetry Protocol)
  - **Storage**: Armazena traces em memória (in-memory)
  - **Query API**: API para consulta de traces
  - **UI Web**: Interface visual para análise de traces

#### 📈 **Metrics (Métricas)**
- **Prometheus** (porta 9090): Sistema de monitoramento e time-series database
  - Coleta métricas HTTP dos serviços via scraping
  - Armazena métricas em TSDB (Time Series Database)
  - Query API para consultas PromQL
  - UI para exploração de métricas

#### 📝 **Logs (Agregação de Logs)**
- **Loki** (porta 3100): Sistema de agregação de logs inspirado no Prometheus
  - Armazena logs de forma eficiente
  - Indexação apenas de labels (não do conteúdo)
  - Query API para consultas LogQL
  
- **Promtail**: Agente de coleta de logs
  - Tail de logs dos containers Docker
  - Service Discovery automático
  - Adiciona labels automaticamente
  - Push de logs para Loki

#### 🎨 **Visualization (Visualização Unificada)**
- **Grafana** (porta 3000): Plataforma de visualização e analytics
  - Dashboards pré-configurados
  - Integração com 3 datasources: Jaeger, Prometheus, Loki
  - Correlação automática entre traces, métricas e logs
  - Auto-refresh e alerting

## 🚀 Como executar

### 1️⃣ **Iniciar todos os serviços:**
```bash
docker-compose up --build
```

Aguarde até ver as mensagens:
```
✅ OpenTelemetry inicializado com sucesso!
🚀 Gateway Service rodando na porta 8080...
🚀 Products Service rodando na porta 8081...
```

### 2️⃣ **Testar a aplicação:**
```bash
# Fazer algumas requisições para gerar traces
curl http://localhost:8080/products

# Ou gerar múltiplas requisições
for i in {1..20}; do curl http://localhost:8080/products; sleep 0.5; done
```

### 3️⃣ **Acessar as interfaces de observabilidade:**

| Interface | URL | Credenciais | Descrição |
|-----------|-----|-------------|-----------|
| **Jaeger UI** | http://localhost:16686 | - | Visualizar traces distribuídos |
| **Prometheus** | http://localhost:9090 | - | Explorar métricas |
| **Grafana** | http://localhost:3000 | admin/admin | Dashboards unificados |

## 🔍 Utilizando o Jaeger UI

### **Básico:**
1. Acesse http://localhost:16686
2. No dropdown **"Service"**, selecione `gateway-service` ou `products-service`
3. Clique em **"Find Traces"**
4. Clique em um trace específico para ver detalhes completos

### **O que observar nos traces:**

- **🎯 Spans**: Cada operação gera um span (segmento de tempo)
  - `gateway.getProducts`: Span principal do gateway
  - `http.call.products-service`: Chamada HTTP entre serviços
  - `products.getAll`: Processamento no serviço de produtos
  - `database.query`: Simulação de query no banco (100ms)
  - `products.filter`: Simulação de filtro (30ms)

- **⏱️ Duração**: Tempo de execução de cada operação
  - Visualização em timeline
  - Identificação de gargalos
  - Análise de latência

- **🏷️ Tags**: Metadados estruturados adicionados aos spans
  - `http.method`: GET, POST, etc.
  - `http.url`: URL completa da requisição
  - `http.status_code`: 200, 404, 500, etc.
  - `service.name`: Nome do serviço
  - `db.system`: postgresql (simulado)
  - `db.statement`: SQL query (simulado)

- **📋 Logs/Events**: Eventos registrados durante a execução
  - "fetching products from database"
  - "filtering products"
  - "products processed"
  - "response sent successfully"

- **🔗 Relações**: Como os spans se relacionam
  - Parent-child: Hierarquia de chamadas
  - Propagação de contexto entre serviços
  - Trace completo end-to-end

### **Recursos Avançados:**
- **Dependency Graph**: Visualiza dependências entre serviços
- **Compare Traces**: Compara múltiplos traces lado a lado
- **Deep Linking**: Link direto para um trace específico
- **System Architecture**: Mapa de arquitetura gerado automaticamente

## 📊 Utilizando o Prometheus

1. Acesse http://localhost:9090
2. Clique em **"Graph"**
3. Teste estas queries:

```promql
# Taxa de requisições por segundo
rate(promhttp_metric_handler_requests_total[1m])

# Uso de memória
go_memstats_alloc_bytes

# Goroutines ativas
go_goroutines

# Serviços UP
up{job=~"gateway|products"}
```

## 🎨 Utilizando o Grafana

### **Acessar:**
1. Abra http://localhost:3000
2. Login: `admin` / Senha: `admin`
3. (Opcional) Troque a senha ou clique "Skip"

### **Dashboards Disponíveis:**

#### 📈 **Microservices Overview**
- Visão geral de todos os serviços
- Request rate por serviço
- Status (UP/DOWN)
- Uso de memória e goroutines
- Logs em tempo real

#### 📊 **Services Detail Metrics**
- Métricas detalhadas de performance
- HTTP request rate e total de requisições
- Uso detalhado de memória (allocated vs system)
- Goroutines e threads ativos
- Taxa de Garbage Collection

#### 🔍 **Observability Dashboard (OpenTelemetry + Jaeger)**
- Foco em observabilidade moderna
- Status dos serviços rastreados
- Request rate e totais
- Logs dos serviços
- Link direto para Jaeger UI

### **Explorando Correlações:**
Grafana permite correlacionar dados dos 3 pilares:

1. **Ver um trace no Jaeger** → Identificar timestamp
2. **Buscar métricas no Prometheus** → Ver CPU/memória naquele momento
3. **Buscar logs no Loki** → Ver erros/warnings relacionados

## 🔄 Estrutura de um Trace Completo

Quando você faz `curl http://localhost:8080/products`, o trace mostra:

```
📍 Trace ID: abc123...
├─ 🌐 gateway.getProducts [200ms total]
│  ├─ Tags:
│  │  ├─ http.method: GET
│  │  ├─ http.url: /products
│  │  └─ service: gateway-service
│  │
│  ├─ Event: "Processando requisição GET /products"
│  │
│  └─ 📡 http.call.products-service [150ms]
│     ├─ Tags:
│     │  ├─ http.url: http://products:8081/products
│     │  └─ peer.service: products-service
│     │
│     └─ 📦 products.getAll [140ms]
│        ├─ Tags:
│        │  ├─ http.method: GET
│        │  ├─ products.count: 4
│        │  └─ service: products-service
│        │
│        ├─ 🗄️ database.query [100ms]
│        │  ├─ Tags:
│        │  │  ├─ db.system: postgresql
│        │  │  ├─ db.statement: SELECT * FROM products
│        │  │  └─ db.operation: SELECT
│        │  └─ Event: "fetching products from database"
│        │
│        ├─ 🔍 products.filter [30ms]
│        │  ├─ Tags:
│        │  │  └─ filter.type: price-range
│        │  └─ Event: "filtering products"
│        │
│        └─ Event: "response sent successfully"
```

## 💡 Conceitos Importantes

### **OpenTelemetry (Moderno)**
- **Padrão CNCF**: Cloud Native Computing Foundation standard
- **Vendor-neutral**: Funciona com Jaeger, Zipkin, Datadog, etc.
- **OTLP Protocol**: OpenTelemetry Protocol (HTTP/gRPC)
- **SDK**: Biblioteca única para traces, metrics e logs
- **Auto-instrumentation**: Propagação automática de contexto

### **Distributed Tracing**
- **Trace**: Representa uma requisição completa através de todos os serviços
- **Span**: Representa uma operação individual dentro de um trace
  - Server Span: Recebe requisição
  - Client Span: Faz requisição externa
  - Internal Span: Operação interna (DB, cache, etc.)
- **Trace ID**: Identificador único do trace (propagado entre serviços)
- **Span ID**: Identificador único do span

### **Context Propagation**
Como o contexto do trace é passado entre serviços:
```
Gateway Service
  └─ HTTP Request Headers:
      ├─ traceparent: 00-{trace-id}-{span-id}-01
      └─ tracestate: ...
          ↓
     Products Service (extrai contexto e continua o trace)
```

### **Três Pilares da Observabilidade**
1. **📊 Metrics (Métricas)**: O QUE está acontecendo
   - Request rate, latência, erro rate
   - Métricas de sistema (CPU, memória)
   
2. **🔍 Traces (Rastreamento)**: ONDE está o problema
   - Qual serviço está lento?
   - Qual operação falhou?
   - Dependências entre serviços
   
3. **📝 Logs (Registros)**: POR QUE aconteceu
   - Mensagens de erro detalhadas
   - Stack traces
   - Contexto da aplicação

### **Sampling**
- **AlwaysSample**: 100% dos traces são coletados (usado neste projeto)
- **ProbabilitySample**: Amostra probabilística (ex: 10%)
- **RateLimiting**: Limite de traces por segundo

### **Tags vs Logs vs Events**
- **Tags**: Metadados estruturados (indexados, queryable)
- **Logs**: Eventos temporais com timestamp
- **Events**: Alias para logs no OpenTelemetry

## 🔧 Tecnologias Utilizadas

### **Backend**
- **Go 1.23**: Linguagem de programação
- **OpenTelemetry SDK v1.32.0**: Instrumentação moderna
- **OTLP Exporter**: Exportador HTTP para Jaeger
- **Prometheus Client**: Exposição de métricas
- **net/http**: Servidor HTTP padrão

### **Observabilidade**
- **Jaeger**: Distributed tracing (CNCF project)
- **Prometheus**: Metrics & monitoring (CNCF project)
- **Loki**: Log aggregation
- **Promtail**: Log collector
- **Grafana**: Visualization platform

### **Infraestrutura**
- **Docker & Docker Compose**: Containerização
- **Bridge Network**: Comunicação entre containers

## 📁 Estrutura do Projeto

```
.
├── docker-compose.yml           # Orquestração de todos os serviços
├── gateway/
│   ├── Dockerfile
│   ├── go.mod                   # Dependências OpenTelemetry
│   └── main.go                  # Código com OTLP exporter
├── products/
│   ├── Dockerfile
│   ├── go.mod                   # Dependências OpenTelemetry
│   └── main.go                  # Código com OTLP exporter
├── prometheus/
│   └── prometheus.yml           # Config de scraping
├── promtail/
│   └── promtail-config.yml      # Config de coleta de logs
└── grafana/
    └── provisioning/
        ├── datasources/
        │   └── datasources.yml  # Jaeger, Prometheus, Loki
        └── dashboards/
            ├── dashboard.yml
            └── dashboards/      # Dashboards pré-configurados
                ├── overview.json
                ├── services-detail.json
                └── observability-dashboard.json
```

## 🛠️ Endpoints Disponíveis

### **Aplicação**
```bash
# Gateway
curl http://localhost:8080/products   # Lista produtos
curl http://localhost:8080/health     # Health check
curl http://localhost:8080/metrics    # Métricas Prometheus

# Products (interno)
curl http://localhost:8081/products   # Lista produtos
curl http://localhost:8081/health     # Health check
curl http://localhost:8081/metrics    # Métricas Prometheus
```

### **Observabilidade - APIs**
```bash
# Jaeger - Listar serviços
curl http://localhost:16686/api/services

# Jaeger - Buscar traces
curl "http://localhost:16686/api/traces?service=gateway-service&limit=10"

# Prometheus - Query
curl "http://localhost:9090/api/v1/query?query=up"

# Loki - Labels
curl http://localhost:3100/loki/api/v1/labels

# Loki - Query
curl -G "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={container_name="gateway"}'
```

## 🎯 Casos de Uso

### **1. Identificar Gargalos de Performance**
1. Acesse Jaeger UI
2. Encontre traces com alta latência
3. Analise qual span está demorando mais
4. Identifique o serviço/operação problemática

### **2. Debug de Erros em Produção**
1. Veja erro nos logs (Grafana → Loki)
2. Identifique o timestamp do erro
3. Busque o trace correspondente no Jaeger
4. Analise toda a cadeia de chamadas
5. Veja tags de erro e stack traces

### **3. Monitoramento de SLOs**
1. Use Prometheus para métricas de SLI
2. Configure alertas no Grafana
3. Correlacione com traces quando alertas disparam
4. Análise de causa raiz com logs

### **4. Análise de Dependências**
1. Use Jaeger Dependency Graph
2. Visualize arquitetura real vs esperada
3. Identifique dependências não documentadas
4. Otimize caminhos críticos

## 🔍 Troubleshooting

### **Serviços não aparecem no Jaeger?**
```bash
# Verifique se serviços estão enviando traces
docker logs gateway 2>&1 | grep -i otel
docker logs products 2>&1 | grep -i otel

# Deve mostrar:
# ✅ OpenTelemetry inicializado com sucesso!

# Verifique API do Jaeger
curl http://localhost:16686/api/services
```

### **Métricas não aparecem no Prometheus?**
```bash
# Verifique targets
curl http://localhost:9090/api/v1/targets

# Teste endpoint de métricas
curl http://localhost:8080/metrics
curl http://localhost:8081/metrics
```

### **Logs não aparecem no Loki?**
```bash
# Verifique Promtail
docker logs promtail 2>&1 | tail -20

# Verifique labels no Loki
curl http://localhost:3100/loki/api/v1/labels
```

### **Dashboard não carrega no Grafana?**
```bash
# Reinicie o Grafana
docker-compose restart grafana

# Verifique logs
docker logs grafana
```

## 🛑 Parar os serviços

```bash
# Para e remove containers
docker-compose down

# Para, remove containers e volumes (perde dados)
docker-compose down -v
```

## 📚 Recursos Adicionais

### **Documentação Oficial**
- [OpenTelemetry](https://opentelemetry.io/docs/)
- [Jaeger](https://www.jaegertracing.io/docs/)
- [Prometheus](https://prometheus.io/docs/)
- [Grafana](https://grafana.com/docs/)
- [Loki](https://grafana.com/docs/loki/)

### **Especificações**
- [OTLP Protocol](https://opentelemetry.io/docs/specs/otlp/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)

### **Tutoriais**
- [OpenTelemetry Go Getting Started](https://opentelemetry.io/docs/languages/go/getting-started/)
- [Jaeger Getting Started](https://www.jaegertracing.io/docs/getting-started/)
- [Prometheus First Steps](https://prometheus.io/docs/introduction/first_steps/)

## 🎓 Observações e Boas Práticas

### **Neste Projeto de Demonstração:**
- ✅ 100% sampling (todos os traces coletados)
- ✅ Latências simuladas (100ms DB, 30ms filtro)
- ✅ Armazenamento in-memory (dados perdidos ao reiniciar)
- ✅ Single-node deployment (todos serviços em um host)

### **Em Produção, Considere:**
- 🎯 **Sampling inteligente**: 1-10% dos traces
- 💾 **Storage persistente**: Elasticsearch, Cassandra
- 🔒 **Segurança**: Autenticação, TLS, RBAC
- 📊 **High availability**: Múltiplas réplicas
- ⚡ **Performance**: Async exporters, batching
- 🔔 **Alerting**: Integração com PagerDuty, Slack
- 📈 **Retention policies**: Retenção de dados configurável
- 🌍 **Service mesh**: Istio/Linkerd para observabilidade automática

## 🚀 Próximos Passos

Para expandir este projeto:

1. **Adicionar Redis** para demonstrar cache tracing
2. **Adicionar PostgreSQL** com queries reais rastreadas
3. **Implementar Circuit Breaker** com spans de fallback
4. **Adicionar autenticação** e rastrear token propagation
5. **Implementar rate limiting** com traces
6. **Adicionar filas** (RabbitMQ/Kafka) com async tracing
7. **Deploy em Kubernetes** com service mesh
8. **Adicionar testes** de integração com tracing

---

**Desenvolvido para demonstrar observabilidade moderna com OpenTelemetry** 🔭

Dúvidas? Abra uma issue ou consulte a documentação oficial do OpenTelemetry!
