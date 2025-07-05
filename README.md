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
