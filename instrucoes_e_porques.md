# Projeto Korp - Instruções de Execução e Justificativas Técnicas

Este documento reúne o passo a passo completo para execução e re-deploy manual da aplicação, assim como o embasamento e justificativa de todas as escolhas técnicas, de arquitetura e dos recursos provisionados na AWS.

---

## 1. Instruções de Execução e Deploy Manual

A infraestrutura foi provisionada na conta AWS (Profile `ITPREMIUM`) em uma instância EC2 Ubuntu 24.04 LTS (`98.92.196.222`).

Você tem duas formas de executar o deploy manualmente:

### Método 1: Via Ansible (Automação 100% Idempotente)

Recomendado para garantir que a aplicação, configurações e dependências estejam sempre sincronizadas no servidor remoto.

1. **Acesse a pasta `ansible` do projeto**:
   ```bash
   cd ~/labs/Projeto_Korp/ansible
   ```

2. **Ative o ambiente virtual Python** (onde o Ansible está instalado):
   ```bash
   source ~/.gemini/antigravity-cli/brain/8efcb02c-6351-4f82-b9c0-61e79d0a2d07/scratch/venv/bin/activate
   ```

3. **Execute o playbook**:
   ```bash
   ansible-playbook -i inventory.ini playbook.yml
   ```

---

### Método 2: Via SSH e Docker Compose (Direto no Servidor AWS)

Útil para debugar, verificar status ou realizar alterações manuais rápidas.

1. **Conecte-se à instância EC2 via SSH**:
   ```bash
   ssh -i ~/.ssh/korp-ec2-key.pem ubuntu@98.92.196.222
   ```

2. **Navegue até o diretório da aplicação**:
   ```bash
   cd /opt/projeto-korp
   ```

3. **Reconstrua e inicie a stack**:
   ```bash
   sudo docker compose up -d --build
   ```

4. **Verifique os containers em execução**:
   ```bash
   sudo docker compose ps
   ```

5. **Acompanhe os logs em tempo real**:
   ```bash
   sudo docker compose logs -f
   ```

---

## 2. Endpoints e URLs de Acesso Público

- **Aplicação Go (via Nginx Proxy Reverso - Porta 80)**:
  - URL: [http://98.92.196.222/projeto-korp](http://98.92.196.222/projeto-korp)
  - Retorno Esperado (JSON):
    ```json
    {
      "nome": "Projeto Korp",
      "horario": "2026-08-18T19:30:08Z"
    }
    ```

- **Grafana Dashboard (Porta 3000)**:
  - URL: [http://98.92.196.222:3000](http://98.92.196.222:3000)
  - Credenciais: `admin` / `admin`

- **Prometheus (Porta 9090)**:
  - URL: [http://98.92.196.222:9090](http://98.92.196.222:9090)

---

## 3. Justificativas das Escolhas Técnicas ("Os Porquês")

### A. Tecnologia da Aplicação: Go (Golang)
- **Performance e Eficiência**: Go gera um binário estático único e leve sem necessidade de uma virtual machine pesada (como JVM ou Node.js runtime).
- **Consumo Mínimo de Recursos**: Ideal para instâncias `t3.micro`, consumindo poucos megabytes de memória RAM.
- **Métricas Nativas**: Utiliza a biblioteca oficial `github.com/prometheus/client_golang` para expor o endpoint `/metrics` com métricas padrão do processo Go e contadores de requisições HTTP.

### B. Containerização com Docker Multi-stage
- **Dockerfile Multi-Stage**:
  - **Stage 1 (Build)**: Usa `golang:1.23-alpine` para compilar estaticamente o binário com `CGO_ENABLED=0`.
  - **Stage 2 (Runtime)**: Usa `alpine:3.20` contendo apenas o binário final, reduzindo drasticamente a superfície de ataque e o tamanho da imagem (para menos de 20MB).

### C. Servidor Web / Proxy Reverso: NGINX
- **Isolamento de Segurança**: A aplicação Go não fica exposta diretamente na porta 80. O NGINX recebe o tráfego HTTP na porta 80 e faz o repasse (`proxy_pass`) de forma transparente para a porta 8080 interna dos containers.
- **Resiliência e Escalabilidade**: O NGINX lida com o gerenciamento de conexões HTTP e cabeçalhos de forma otimizada.

### D. Observabilidade e Monitoramento: Prometheus + Grafana
- **Prometheus**: Faz scraping periódico (a cada 5s) do endpoint `/metrics` exportado pela aplicação Go, registrando o volume de requisições e a saúde do serviço.
- **Grafana com Provisionamento por Código (IaC)**: Os datasources e dashboards são carregados automaticamente via arquivos YAML e JSON no boot, eliminando a necessidade de configuração manual pela interface gráfica.

### E. Automação e Gerenciamento de Configuração: Ansible
- **Sem Necessidade de Agente Remote**: Opera diretamente via SSH (`agentless`).
- **Idempotência**: Garante que a infraestrutura atinja e permaneça no estado desejado independentemente de quantas vezes o playbook for reexecutado.

### F. Recursos Provisionados na AWS (`ITPREMIUM`)
- **EC2 `t3.micro` (Ubuntu 24.04 LTS)**: Excelente custo-benefício, elegível para AWS Free Tier e suficiente para suportar toda a stack containerizada em Docker.
- **Security Group `projeto-korp-sg`**:
  - Libera apenas as portas estritamente necessárias: `22` (SSH gerenciado), `80` (Aplicação via Nginx), `3000` (Grafana) e `9090` (Prometheus).
- **Chave SSH (`korp-ec2-key.pem`)**: Gerada e associada exclusivamente para acesso seguro via terminal e Ansible.
