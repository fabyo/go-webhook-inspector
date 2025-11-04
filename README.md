# <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go"/> Golang Webhook Inspector

![HTTP](https://img.shields.io/badge/HTTP-inspector-green?style=for-the-badge)
![Status](https://img.shields.io/badge/status-experimental-orange?style=for-the-badge)

Um pequeno **inspector de webhooks em Go**.  
Ele expõe endpoints HTTP para receber requisições (por exemplo, webhooks de **gateways** de **pagamento**, ERPs, APIs externas), registra os dados em memória e oferece uma API para inspecionar cada requisição recebida em detalhe.

---

## ✅ Objetivo

Fornecer um serviço simples em Go para:

- Capturar requisições HTTP (especialmente webhooks).
- Visualizar o payload real enviado por serviços externos.
- Ajudar no desenvolvimento e depuração de integrações entre sistemas.

É um utilitário útil tanto para testes locais quanto como base para uma API mais completa.

---

## 🔗 Endpoints

### `POST /hook`

Endpoint de captura.

Qualquer sistema pode enviar requisições HTTP para este caminho.  
O servidor registra:

- Método HTTP
- Path
- Headers
- Corpo da requisição (body)
- IP de origem
- Data/hora da recepção

Exemplo usando `curl`:

```bash
curl -X POST http://localhost:8082/hook   -H "Content-Type: application/json"   -d '{"pedido":123,"cliente":"Fabyo","valor":199.90,"status":"pago"}'
```

Resposta (exemplo):

```json
{
  "id": 1,
  "method": "POST",
  "path": "/hook",
  "headers": {
    "Content-Type": "application/json",
    "User-Agent": "curl/8.4.0"
  },
  "body": "{\"pedido\":123,\"cliente\":\"Fabyo\",\"valor\":199.9,\"status\":\"pago\"}",
  "remote_ip": "127.0.0.1:54321",
  "created_at": "2025-11-04T18:23:45.123456789Z"
}
```

---

### `GET /events`

Retorna a lista dos últimos eventos registrados em `/hook`.

```bash
curl http://localhost:8082/events
```

Exemplo de resposta:

```json
[
  {
    "id": 1,
    "method": "POST",
    "path": "/hook",
    "headers": { "...": "..." },
    "body": "{...}",
    "remote_ip": "127.0.0.1:54321",
    "created_at": "2025-11-04T18:23:45.123456789Z"
  }
]
```

---

### `GET /events/{id}`

Retorna os detalhes de um evento específico.

```bash
curl http://localhost:8082/events/1
```

---

## 🧱 Estrutura básica

O servidor mantém os eventos em memória, usando uma estrutura simples:

- `Event` – representa uma requisição recebida.
- `Store` – armazena os eventos em slice e usa `sync.Mutex` para garantir segurança em ambiente concorrente.
- Limite configurável de eventos em memória (por padrão, mantém apenas os últimos N).

Não há banco de dados neste projeto por padrão.  
A ideia é ser leve, simples e focado em desenvolvimento/local.

---

## ⚙️ Como rodar

### Pré-requisitos

- Go 1.20+ instalado no sistema.

### Passos

1. Clonar o repositório:

```bash
git clone https://github.com/fabyo/go-webhook-inspector.git
cd go-webhook-inspector
```

2. Inicializar (se necessário) e baixar dependências:

```bash
go mod tidy
```

3. Rodar o servidor:

```bash
go run main.go
```

O servidor ficará disponível em:

```text
http://localhost:8082
```

---

## 🧪 Testando rapidamente

### 1. Ver mensagem inicial

Abra no navegador:

```text
http://localhost:8082/
```

### 2. Enviar uma requisição de teste

```bash
curl -X POST http://localhost:8082/hook   -H "Content-Type: application/json"   -d '{"teste":"ok"}'
```

### 3. Listar eventos

```bash
curl http://localhost:8082/events
```

### 4. Ver um evento específico

```bash
curl http://localhost:8082/events/1
```

---

## 💡 Ideias de evolução

- Persistir eventos em:
  - Arquivo local
  - SQLite / PostgreSQL / Redis
- Adicionar uma interface web (HTML/JS) para visualizar os eventos em tempo real.
- Filtros por header, path, método, intervalo de datas.
- Suporte a autenticação para proteger os endpoints (token, basic auth, etc.).
- Integração com tunelamento (ex.: ngrok) para receber webhooks de serviços externos diretamente no ambiente local.

---

## 📜 Licença

Escolha e adicione aqui a licença desejada (MIT, Apache 2.0, etc.).
