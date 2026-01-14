# Exemplo de Tracing com Jaeger

Este projeto demonstra o uso do Jaeger para rastreamento distribuído entre microsserviços.

## Arquitetura

- **Gateway Service** (porta 8080): Serviço que recebe requisições externas
- **Products Service** (porta 8081): Serviço que gerencia produtos
- **Jaeger** (porta 16686): Interface UI para visualização de traces

## Como executar

1. **Iniciar os serviços:**
```bash
docker-compose up --build
```

2. **Testar a aplicação:**
```bash
curl http://localhost:8080/api/products
```

3. **Visualizar os traces no Jaeger:**
Abra o navegador em: http://localhost:16686

## Utilizando o Jaeger UI

1. Acesse http://localhost:16686
2. Selecione o serviço "gateway-service" ou "products-service" no dropdown
3. Clique em "Find Traces"
4. Clique em um trace específico para ver detalhes

### O que observar nos traces:

- **Spans**: Cada operação gera um span (segmento de tempo)
- **Duração**: Tempo de execução de cada operação
- **Tags**: Metadados adicionados (HTTP status, URLs, etc.)
- **Logs**: Eventos registrados durante a execução
- **Relações**: Como os serviços se comunicam (parent-child spans)

## Estrutura do Trace

Quando você faz uma requisição para `/api/products`, o trace mostra:

1. **gateway-get-products** (Gateway)
   - Validação da requisição
   - Chamada HTTP para o serviço de produtos
   
2. **products-get-all** (Products)
   - Busca simulada no banco de dados
   - Filtro/processamento dos produtos
   - Retorno dos dados

## Conceitos importantes

- **Trace**: Representa uma requisição completa através de todos os serviços
- **Span**: Representa uma operação individual dentro de um trace
- **Context Propagation**: Como o contexto do trace é passado entre serviços (via HTTP headers)
- **Tags**: Metadados estruturados (ex: http.status_code=200)
- **Logs**: Eventos temporais dentro de um span

## Parar os serviços

```bash
docker-compose down
```

## Observações

- Os serviços simulam latências para facilitar visualização
- Todos os traces são amostrados (100% sampling)
- Logs aparecem no console e no Jaeger