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
