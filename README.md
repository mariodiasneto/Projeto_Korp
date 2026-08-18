# Desafio Técnico - Projeto Korp

Este repositório contém a solução completa para o desafio técnico da Korp, contemplando o desenvolvimento do serviço HTTP em Golang, containerização com Docker & NGINX Reverse Proxy, observabilidade com Prometheus + Grafana e automação de infraestrutura via Ansible para deploy em ambiente Linux (AWS EC2).

---

## 🛠️ Arquitetura e Recursos Necessários (Justificativa das Escolhas)

### 1. Aplicação (`http-server-projeto-korp`)
- **Linguagem**: Golang (1.23)
- **Justificativa**:
  - Go possui alto desempenho com baixo consumo de memória e tempo de inicialização instantâneo.
  - Utiliza o pacote nativo `net/http` e a biblioteca oficial `github.com/prometheus/client_golang` para exposição estrita de métricas.
  - O horário é retornado em UTC formatado dinamicamente via `time.Now().UTC().Format(time.RFC3339)`.
- **Dockerfile (Multi-Stage)**:
  - `golang:1.23-alpine` (builder) -> `alpine:3.20` (runner).
  - Reduz drasticamente o tamanho final da imagem (~15MB) e minimiza a superfície de ataque em produção.

### 2. Rede & Proxy Reverso (`NGINX`)
- **Imagem**: `nginx:alpine`
- **Mapeamento de Portas**: Porta 80 do host -> Porta 80 do container NGINX.
- **Rede Docker**: `korp-network` (Modo `bridge`).
- **Justificativa**:
  - O container da aplicação Golang **não expõe portas diretamente para o host**, garantindo isolamento de rede.
  - O NGINX atua como camada de proxy reverso, gerenciando o tráfego de entrada na porta 80 e encaminhando para `http://http-server-projeto-korp:8080`.

### 3. Observabilidade e Monitoramento (`Prometheus` + `Grafana`)
- **Prometheus (`prom/prometheus`)**:
  - Coleta métricas a cada 5 segundos do endpoint `/metrics` exposto pelo serviço Go.
  - Métricas obrigatórias implementadas:
    - `http_server_up`: Gauge indicando disponibilidade (1 = UP, 0 = DOWN).
    - `http_requests_total`: Counter registrando volume total e taxa de requisições agrupadas por caminho, método e status.
- **Grafana (`grafana/grafana`)**:
  - Provisionamento automatizado via arquivos de configuração (`datasource.yml`, `dashboard.yml` e `http-server-projeto-korp-dashboard.json`).
  - Dashboard pronto acessível na porta `3000` sem necessidade de configuração manual.

### 4. Automação e Provisionamento (`Ansible`)
- **Playbook (`ansible/playbook.yml`)**:
  - Instala o Docker e Docker Compose na máquina de destino (Ubuntu/Debian ou RHEL/Amazon Linux).
  - Garante a criação do diretório `/opt/projeto-korp` e sincronização dos arquivos do projeto.
  - Executa o build da imagem Docker e sobe todos os containers (`docker compose up -d --build`).
  - Valida o funcionamento do endpoint `GET /projeto-korp` via requisição HTTP NGINX e exibe o payload JSON retornado no console.

### 5. Recursos Recomendados na AWS (para Teste de Deploy)
- **Instância EC2**: `t3.micro` ou `t2.micro` (Elegível para o AWS Free Tier, 1 vCPU, 1 GB RAM).
- **Sistema Operacional**: Ubuntu 24.04 LTS ou Amazon Linux 2023.
- **Security Group (Grupo de Segurança)**:
  - `SSH (22)`: Acesso administrativo via Ansible.
  - `HTTP (80)`: Acesso ao serviço via NGINX Proxy Reverso.
  - `Grafana (3000)`: Visualização do Dashboard de monitoramento.
  - `Prometheus (9090)`: (Opcional) Acesso direto à interface do Prometheus.

---

## 📁 Estrutura de Pastas do Projeto

```
Projeto_Korp/
├── app/
│   ├── main.go                     # Código-fonte da aplicação HTTP Go com métricas Prometheus
│   ├── go.mod                      # Módulo Go e dependências
│   └── Dockerfile                  # Multi-stage Dockerfile para compilação e execução
├── nginx/
│   └── http-server-projeto-korp.conf  # Configuração do Proxy Reverso NGINX (porta 80 -> 8080)
├── prometheus/
│   └── prometheus.yml              # Configuração de coleta (scrape) do Prometheus
├── grafana/
│   └── provisioning/
│       ├── datasources/
│       │   └── datasource.yml      # Data Source do Prometheus no Grafana
│       └── dashboards/
│           ├── dashboard.yml       # Provedor de dashboards automatizado
│           └── http-server-projeto-korp-dashboard.json  # Dashboard pré-configurado
├── ansible/
│   ├── ansible.cfg                 # Configuração padrão do Ansible
│   ├── inventory.ini               # Inventário de servidores
│   └── playbook.yml                # Playbook de automação total do ambiente
├── docker-compose.yml              # Orquestração dos 4 containers (Go App, Nginx, Prometheus, Grafana)
├── instrucoes_para_a_realizacao_do_desafio.md
└── README.md
```

---

## 🚀 Como Executar Localmente (Docker Compose)

Para testar rapidamente o ambiente em sua máquina local via Docker Compose:

1. **Subir os containers**:
   ```bash
   docker compose up -d --build
   ```

2. **Testar o endpoint da aplicação (via NGINX)**:
   ```bash
   curl -i http://localhost:80/projeto-korp
   ```
   **Resposta Esperada**:
   ```json
   {
     "nome": "Projeto Korp",
     "horario": "2026-08-18T17:30:00Z"
   }
   ```

3. **Acessar as interfaces de monitoramento**:
   - **Grafana Dashboard**: `http://localhost:3000` (Usuário: `admin`, Senha: `admin`)
   - **Prometheus UI**: `http://localhost:9090`

---

## ☁️ Deploy na AWS com Ansible

### Passo 1: Provisionar a Instância EC2 na AWS
Crie uma instância EC2 (ex: `t3.micro` Ubuntu 24.04) com o Security Group configurado para as portas **22**, **80** e **3000**.

### Passo 2: Configurar o Inventário do Ansible
Edite o arquivo `ansible/inventory.ini` inserindo o IP público da sua instância AWS e o caminho da sua chave privada SSH (`.pem`):

```ini
[app_servers]
aws-korp ansible_host=SEU_IP_PUBLICO_AWS ansible_user=ubuntu ansible_ssh_private_key_file=~/.ssh/sua-chave.pem
```

### Passo 3: Executar o Playbook Ansible
No seu terminal local, execute:

```bash
cd ansible
ansible-playbook -i inventory.ini playbook.yml
```

O Ansible irá:
1. Instalar o Docker & Docker Compose na EC2.
2. Transferir os arquivos do projeto.
3. Fazer o build e iniciar a stack completa.
4. Validar o funcionamento fazendo requisição HTTP em `http://localhost:80/projeto-korp` dentro do servidor e exibir a resposta no seu console!
