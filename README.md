# swagger_guard

Auditoria automática de segurança para APIs baseada em especificações OpenAPI (Swagger).

## Objetivo

Fornecer uma ferramenta CLI em Go para analisar especificações OpenAPI, mapeando endpoints, métodos, parâmetros e requisitos de segurança, facilitando a auditoria e identificação de potenciais riscos.

## Instalação

```sh
git clone https://github.com/lorenaziviani/swagger_guard.git
cd swagger_guard
go mod tidy
```

## Uso

Execute o comando parse informando o arquivo da especificação:

```sh
go run main.go parse --file ./api-spec.yaml
```

## Exemplo de api-spec.yaml

```yaml
openapi: 3.0.0
info:
  title: Exemplo de API
  version: "1.0.0"
paths:
  /usuarios:
    get:
      summary: Lista usuários
      responses:
        "200":
          description: OK
    post:
      summary: Cria usuário
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                nome:
                  type: string
      responses:
        "201":
          description: Criado
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-KEY
security:
  - ApiKeyAuth: []
```

## Diagrama

Veja o fluxo inicial do CLI e parser em `docs/cli_openapi_parser.drawio`.

## Checagens automáticas OWASP Top 10

O comando `parse` agora realiza checagens automáticas de segurança baseadas no OWASP Top 10:

- **Rotas sem autenticação** (security: [])
- **Uso de GET para criação/deleção**
- **Ausência de HTTPS** (servers.url não usa https)
- **Parâmetros de query sem tipo**
- **Métodos inseguros** (TRACE, OPTIONS abertos)

Exemplo de saída:

```
OWASP Top 10 Issues:

[No Authentication]
- GET /usuarios

[No HTTPS]
- http://api.exemplo.com

[Query parameter without type]
- GET /usuarios param: filtro
```

Se nenhuma falha for encontrada:

```
No OWASP Top 10 issues found!
```

## Formatos de saída e relatórios

O comando `parse` suporta múltiplos formatos de saída:

- `--output cli` (padrão): saída colorida no terminal
- `--output json`: saída estruturada em JSON
- `--output markdown`: saída em Markdown
- `--output-file <arquivo>`: salva o relatório em arquivo externo

### Exemplo de uso

```sh
go run main.go parse --file ./api-spec.yaml --output cli

go run main.go parse --file ./api-spec.yaml --output json --output-file report.json

go run main.go parse --file ./api-spec.yaml --output markdown --output-file report.md
```

### Severidade das falhas

- **high**: vermelho (No Authentication, Insecure HTTP Methods, No HTTPS)
- **medium**: amarelo (GET used for create/delete)
- **low**: amarelo (Query parameter without type)

### Exemplo de saída CLI

```sh
OWASP Top 10 Issues:

[No Authentication] (HIGH)
- GET /users

[No HTTPS] (HIGH)
- http://api.insegura.com

[Query parameter without type] (LOW)
- GET /users param: filter
```

### Exemplo de saída JSON

```json
{
  "issues": [
    {
      "category": "No Authentication",
      "severity": "high",
      "item": "GET /users"
    },
    {
      "category": "No HTTPS",
      "severity": "high",
      "item": "http://api.insegura.com"
    },
    {
      "category": "Query parameter without type",
      "severity": "low",
      "item": "GET /users param: filter"
    }
  ],
  "summary": { "high": 2, "medium": 0, "low": 1 }
}
```

## Integração com CI/CD

A ferramenta retorna exit code 1 se encontrar falhas de severidade **high** (críticas), permitindo uso em pipelines, pre-commit hooks e GitHub Actions.

### Exemplo de uso em GitHub Actions

```yaml
name: Swagger Guard Scan
on:
  push:
    paths:
      - "api-spec.yaml"
  pull_request:
    paths:
      - "api-spec.yaml"

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.21"
      - name: Install dependencies
        run: go mod tidy
      - name: Run Swagger Guard
        run: go run main.go parse --file api-spec.yaml --output cli
```

Se houver falhas críticas, o job irá falhar (exit code 1).

Também pode ser usado em pre-commit hooks ou outros pipelines CI/CD.

## Métricas de uso

A CLI armazena estatísticas de uso localmente em um banco SQLite (`metrics.db` por padrão):

- Total de execuções
- Total de falhas encontradas por severidade
- Data/hora da última execução

### Como visualizar métricas

```sh
go run main.go parse --metrics
```

### Exemplo de saída

```
==== CLI Metrics ====
Total executions: 5
Total high severity issues: 7
Total medium severity issues: 2
Total low severity issues: 3
Last run: 2024-06-10T15:30:00Z
```

Você pode customizar o caminho do banco com `--metrics-db <arquivo>`.
