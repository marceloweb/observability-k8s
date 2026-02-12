# 🔭 Observability Stack - Kubernetes Helm Chart

![Kubernetes](https://img.shields.io/badge/kubernetes-%23326ce5.svg?style=for-the-badge&logo=kubernetes&logoColor=white)
![Grafana](https://img.shields.io/badge/grafana-%23F46800.svg?style=for-the-badge&logo=grafana&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=for-the-badge&logo=Prometheus&logoColor=white)

## 📖 Sobre o Projeto

Stack completa de observabilidade para microserviços, implementando os três pilares fundamentais:

- **📊 Métricas**: Prometheus + Alertmanager
- **📝 Logs**: Loki + Promtail
- **🔍 Traces**: Tempo + OpenTelemetry Collector

Inclui Grafana como plataforma unificada de visualização e dois microserviços de exemplo (Gateway e Products) já instrumentados com OpenTelemetry.

## ✨ Componentes

### Observabilidade

| Componente      | Versão  | Porta | Descrição                        |
|-----------------|---------|-------|----------------------------------|
| Grafana         | latest  | 3000  | Dashboards e visualização        |
| Prometheus      | latest  | 9090  | Coleta e armazenamento métricas  |
| Alertmanager    | latest  | 9093  | Gerenciamento de alertas         |
| Loki            | latest  | 3100  | Agregação de logs                |
| Promtail        | latest  | -     | Coleta de logs (DaemonSet)       |
| Tempo           | 2.9.0   | 3200  | Distributed tracing              |
| OTel Collector  | latest  | 4317  | Coleta telemetria OpenTelemetry  |

### Aplicações para testes

| Serviço  | Porta | Descrição                    |
|----------|-------|------------------------------|
| Gateway  | 8080  | API Gateway instrumentado    |
| Products | 8081  | Serviço de produtos exemplo  |

## 📋 Pré-requisitos

### Software Necessário

```bash
# Kubernetes Local
- Minikube >= 1.37.0
- kubectl >= 1.34.0

# Package Manager
- Helm >= 4.1.0

```

### Instalação dos Pré-requisitos

#### 🐧 Linux

```bash
# Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

## 🚀 Quick Start

### 1. Iniciar Minikube

```bash
# Iniciar cluster com recursos adequados
minikube start --cpus=4 --memory=8192 --disk-size=20g

# Habilitar addons úteis
minikube addons enable metrics-server
minikube addons enable ingress
```

### 2. Instalar a Stack

```bash
# Clone o repositório
git clone <repository-url>
cd observability-stack

# Instalar usando Helm
helm install obs-stack ./charts/observability-stack \
  -n observability \
  --create-namespace \
  -f values/values-local.yaml

```

### 3. Verificar Instalação

```bash
# Verificar status
helm status obs-stack -n observability

# Ver pods
kubectl get pods -n observability

# Aguardar todos os pods ficarem prontos
kubectl wait --for=condition=ready pod --all -n observability --timeout=300s
```

### 4. Acessar as Interfaces

#### Opção 1: Port-Forward Manual

```bash
# Grafana
kubectl port-forward -n observability svc/grafana 3000:3000

# Prometheus
kubectl port-forward -n observability svc/prometheus 9090:9090

# Gateway (aplicação exemplo)
kubectl port-forward -n observability svc/gateway 8080:8080
```

#### Opção 2: Usando Script

```bash
./scripts/port-forward.sh
```

#### Opção 3: Usando Makefile

```bash
make port-forward
```

### 5. Credenciais Padrão

| Serviço     | URL                   | Usuário | Senha |
|-------------|-----------------------|---------|-------|
| Grafana     | http://localhost:3000 | admin   | admin |
| Prometheus  | http://localhost:9090 | -       | -     |
| Alertmanager| http://localhost:9093 | -       | -     |

## 🎯 Testando a Stack

### Gerar Tráfego de Teste

```bash
# Fazer requisições ao gateway
for i in {1..100}; do
  curl http://localhost:8080/products
  sleep 1
done
```

### Verificar Dados

1. **Métricas** - Prometheus (http://localhost:9090)
   ```promql
   # Requisições HTTP
   rate(http_requests_total[5m])
   
   # Latência P95
   histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
   ```

2. **Logs** - Grafana > Explore > Loki
   ```logql
   {namespace="observability", app="gateway"}
   ```

3. **Traces** - Grafana > Explore > Tempo
   - Buscar por service name: `gateway` ou `products`
   - Ver trace completo da requisição

## 🔧 Configuração

### Personalizar Valores

```bash
# Editar values
vim values/values-local.yaml

# Aplicar mudanças
helm upgrade obs-stack ./charts/observability-stack \
  -n observability \
  -f values/values-local.yaml
```

### Principais Configurações

```yaml
# values.yaml (exemplo)
grafana:
  replicas: 1
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
  persistence:
    enabled: true
    size: 10Gi

prometheus:
  retention: 15d
  resources:
    limits:
      cpu: 500m
      memory: 512Mi

gateway:
  replicas: 2
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5
```

## 📊 Dashboards Grafana

Após instalação, importe os dashboards pré-configurados:

1. Acesse Grafana (http://localhost:3000)
2. Navegue para Dashboards > Browse
3. Os seguintes dashboards estarão disponíveis:
   - **Kubernetes Cluster Monitoring**
   - **Application Metrics**
   - **Logs Dashboard**
   - **Distributed Tracing**

Ou importe manualmente de `docs/dashboards/`

## 🛠️ Comandos Úteis

```bash
# Ver logs de um serviço
kubectl logs -n observability -l app=grafana -f

# Descrever um pod
kubectl describe pod -n observability <pod-name>

# Executar comando em pod
kubectl exec -it -n observability <pod-name> -- /bin/sh

# Ver recursos consumidos
kubectl top pods -n observability

# Reiniciar deployment
kubectl rollout restart deployment/grafana -n observability

# Ver eventos
kubectl get events -n observability --sort-by='.lastTimestamp'
```

## 🔄 Atualização

```bash
# Atualizar chart
helm upgrade obs-stack ./charts/observability-stack \
  -n observability \
  -f values/values-local.yaml

# Ver histórico de releases
helm history obs-stack -n observability

# Rollback se necessário
helm rollback obs-stack <revision> -n observability
```

## 🗑️ Desinstalação

```bash
# Desinstalar release
helm uninstall obs-stack -n observability

# Remover namespace (opcional)
kubectl delete namespace observability

# Ou usando Makefile
make uninstall

# Parar minikube
minikube stop

# Deletar cluster (cuidado!)
minikube delete
```

## 🐛 Troubleshooting

### Pods não iniciam

```bash
# Verificar eventos
kubectl get events -n observability --sort-by='.lastTimestamp'

# Descrever pod com problema
kubectl describe pod -n observability <pod-name>

# Ver logs
kubectl logs -n observability <pod-name>
```

### Problemas de Recursos

```bash
# Verificar recursos do nó
kubectl top nodes

# Verificar recursos dos pods
kubectl top pods -n observability

# Aumentar recursos do Minikube
minikube stop
minikube delete
minikube start --cpus=6 --memory=12288
```

### PVC Pendente

```bash
# Verificar PVCs
kubectl get pvc -n observability

# Verificar StorageClass
kubectl get storageclass

# Se necessário, usar hostPath (apenas local)
# Editar values.yaml e definir storageClass: standard
```

### Grafana sem dados

1. Verificar datasources: Configuration > Data Sources
2. Testar conexão com Prometheus/Loki/Tempo
3. Verificar se serviços estão rodando:
   ```bash
   kubectl get svc -n observability
   ```

Para mais detalhes, consulte [docs/troubleshooting.md](docs/troubleshooting.md)

## 📚 Documentação Adicional

- [Arquitetura Detalhada](docs/architecture.md)
- [Guia de Troubleshooting](docs/troubleshooting.md)
- [Customização de Dashboards](docs/dashboards/)
- [Helm Chart Values](charts/observability-stack/values.yaml)

## 👥 Autor

- Marcelo Lopes Oliveira - [@marceloweb](https://www.linkedin.com/in/marceloweb/)

---

⭐ Se este projeto foi útil, considere dar uma estrela!