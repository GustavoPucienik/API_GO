# System API — Go

Reescrita da API System em Go, usando Chi como roteador e sqlc para acesso ao banco.

## Stack

| Camada | Biblioteca |
|---|---|
| HTTP | `github.com/go-chi/chi/v5` |
| MySQL driver | `github.com/go-sql-driver/mysql` |
| Geração de queries | `sqlc` v1.27 |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Configuração | variáveis de ambiente via `pkg/config` |

## Estrutura

```
cmd/api/main.go                  — entry point
internal/
  admin/                         — módulo admin (auth, usuários, grupos, permissões, menus)
  maintenance/                   — módulo manutenção (ordens, lookups, dashboard)
  middleware/                    — auth.go, permission.go
  db/                            — gerado por sqlc (não editar manualmente)
pkg/
  config/                        — carrega variáveis de ambiente
  jwtutil/                       — geração e validação de tokens
  response/                      — helpers JSON
db/
  schema/admin.sql               — DDL tabelas admin
  schema/maintenance.sql         — DDL tabelas manutenção
  queries/admin.sql              — queries sqlc admin
  queries/maintenance.sql        — queries sqlc manutenção
sqlc.yaml
uploads/orders/                  — arquivos enviados (criado automaticamente)
```

## Configuração

Crie um arquivo `.env` (ou exporte as variáveis):

```
PORT=3000
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASS=senha
DB_NAME=nomedobanco
JWT_SECRET=sua-chave-secreta
```

## Rodar

```bash
go run ./cmd/api
# ou
go build -o systemapi ./cmd/api && ./systemapi
```

Para regenerar o código sqlc após alterar queries:

```bash
sqlc generate
```

---

## Autenticação

Toda rota (exceto `/auth/login`) exige o header:

```
Authorization: Bearer <token>
```

O token é um JWT com payload:

```json
{
  "user_id": 1,
  "name": "João Silva",
  "email": "joao@example.com",
  "exp": 1234567890
}
```

O refresh token tem validade maior e é trocado via `POST /auth/refresh`.

---

## Sistema de Permissões

Cada usuário pertence a grupos, e cada grupo possui `access_level`s. Um `access_level` define, por módulo:

| Campo | Descrição |
|---|---|
| `permission` | Bitmask: READ=1, CREATE=2, UPDATE=4, DELETE=8 |
| `data_scope` | `OWN` / `RESPONSIBLE` / `GROUP` / `ALL` |
| `can_be_assigned` | Se o usuário pode ser atribuído como responsável |

O middleware `Permission(mod, bit)` verifica o bitmask. O `data_scope` mais permissivo entre todos os grupos do usuário é usado para filtrar dados.

---

## Endpoints

### Auth

| Método | Rota | Descrição |
|---|---|---|
| POST | `/auth/login` | Login com email/senha, retorna access + refresh token |
| POST | `/auth/refresh` | Troca refresh token por novo access token |

### Usuários

| Método | Rota | Permissão |
|---|---|---|
| GET | `/user` | READ |
| POST | `/user` | CREATE |
| GET | `/user/{id}` | READ |
| PUT | `/user/{id}` | UPDATE |
| DELETE | `/user/{id}` | DELETE |
| PATCH | `/user/reset-password` | — (autenticado) |
| PATCH | `/user/{id}/reset-password` | UPDATE |

### Grupos

| Método | Rota | Permissão |
|---|---|---|
| GET | `/group` | READ |
| POST | `/group` | CREATE |
| GET | `/group/{id}` | READ |
| PUT | `/group/{id}` | UPDATE |
| DELETE | `/group/{id}` | DELETE |

### Níveis de Acesso

| Método | Rota | Permissão |
|---|---|---|
| GET | `/access-level` | READ |
| POST | `/access-level` | CREATE |
| GET | `/access-level/{id}` | READ |
| PUT | `/access-level/{id}` | UPDATE |
| DELETE | `/access-level/{id}` | DELETE |

### Módulos

| Método | Rota | Permissão |
|---|---|---|
| GET | `/module` | READ |
| POST | `/module` | CREATE |
| GET | `/module/{id}` | READ |
| PUT | `/module/{id}` | UPDATE |
| DELETE | `/module/{id}` | DELETE |

### Menus

| Método | Rota | Permissão |
|---|---|---|
| GET | `/menu` | READ |
| POST | `/menu` | CREATE |
| PUT | `/menu/{id}` | UPDATE |
| DELETE | `/menu/{id}` | DELETE |

---

### Ordens de Manutenção

Módulo: `"Manutenção"`

| Método | Rota | Permissão | Descrição |
|---|---|---|---|
| POST | `/maintenance` | CREATE | Cria nova ordem |
| GET | `/maintenance` | READ | Lista ordens paginadas (com filtros) |
| GET | `/maintenance/{id}` | READ | Busca ordem por ID |
| PUT | `/maintenance/{id}` | UPDATE | Atualiza ordem (campos opcionais) |
| DELETE | `/maintenance/{id}` | DELETE | Remove ordem (soft delete) |
| PATCH | `/maintenance/{id}/assign` | UPDATE | Atribui técnico a uma data de agendamento |
| PATCH | `/maintenance/{id}/signature` | UPDATE | Salva assinatura do cliente (base64) |
| PATCH | `/maintenance/{id}/status` | UPDATE | Altera status (fecha/reabre automaticamente) |

#### Filtros de listagem (`GET /maintenance`)

| Parâmetro | Tipo | Descrição |
|---|---|---|
| `page` | int | Página (default 1) |
| `limit` | int | Itens por página (default 20) |
| `search` | string | Busca em cliente, solicitante, motivo, técnico |
| `statusId` | int | Filtro por status |
| `priorityId` | int | Filtro por prioridade |
| `reasonId` | int | Filtro por motivo |
| `clientCode` | string | Filtro por código do cliente |
| `responsibleId` | int | Filtro por técnico responsável |
| `startDate` | date | Data de criação início (YYYY-MM-DD) |
| `endDate` | date | Data de criação fim (YYYY-MM-DD) |
| `onlyWithScheduleDate` | bool | Apenas ordens com agendamento |

O escopo de dados (`OWN` / `RESPONSIBLE` / `GROUP` / `ALL`) é aplicado automaticamente com base na permissão do usuário.

### Anexos

| Método | Rota | Permissão | Descrição |
|---|---|---|---|
| POST | `/maintenance/{id}/attachments` | CREATE | Upload de arquivos (multipart/form-data, campo `files`) |
| GET | `/maintenance/{id}/attachments` | READ | Lista anexos da ordem |
| DELETE | `/maintenance/{id}/attachments/{attachmentId}` | DELETE | Remove anexo (soft delete + remoção física) |

### Checklists de Ordem

| Método | Rota | Permissão | Descrição |
|---|---|---|---|
| POST | `/maintenance/{id}/checklists` | CREATE | Submete checklist preenchido |
| GET | `/maintenance/{id}/checklists` | READ | Lista checklists da ordem |
| GET | `/maintenance/{id}/checklists/{checklistId}` | READ | Busca checklist por ID |
| DELETE | `/maintenance/{id}/checklists/{checklistId}` | DELETE | Remove checklist |

### Lookups

| Método | Rota | Descrição |
|---|---|---|
| GET | `/maintenance-lookups` | Retorna todos os lookups (status, prioridade, motivo, nota, tipo de tempo) em uma chamada |
| GET | `/maintenance-lookups/status` | Lista status |
| POST | `/maintenance-lookups/status` | Cria status |
| PUT | `/maintenance-lookups/status/{id}` | Atualiza status |
| DELETE | `/maintenance-lookups/status/{id}` | Remove status |
| GET | `/maintenance-lookups/priority` | Lista prioridades |
| POST | `/maintenance-lookups/priority` | Cria prioridade |
| PUT | `/maintenance-lookups/priority/{id}` | Atualiza prioridade |
| DELETE | `/maintenance-lookups/priority/{id}` | Remove prioridade |
| GET | `/maintenance-lookups/reason` | Lista motivos |
| POST | `/maintenance-lookups/reason` | Cria motivo |
| PUT | `/maintenance-lookups/reason/{id}` | Atualiza motivo |
| DELETE | `/maintenance-lookups/reason/{id}` | Remove motivo |
| GET | `/maintenance-lookups/note-template` | Lista modelos de nota |
| POST | `/maintenance-lookups/note-template` | Cria modelo de nota |
| PUT | `/maintenance-lookups/note-template/{id}` | Atualiza modelo |
| DELETE | `/maintenance-lookups/note-template/{id}` | Remove modelo |
| GET | `/maintenance-lookups/time-entry-type` | Lista tipos de apontamento |
| POST | `/maintenance-lookups/time-entry-type` | Cria tipo de apontamento |
| PUT | `/maintenance-lookups/time-entry-type/{id}` | Atualiza tipo |
| DELETE | `/maintenance-lookups/time-entry-type/{id}` | Remove tipo |

### Templates de Checklist

| Método | Rota | Permissão | Descrição |
|---|---|---|---|
| GET | `/maintenance-checklist-template` | READ | Lista templates |
| POST | `/maintenance-checklist-template` | CREATE | Cria template |
| GET | `/maintenance-checklist-template/{id}` | READ | Busca template por ID |
| PUT | `/maintenance-checklist-template/{id}` | UPDATE | Atualiza template |
| DELETE | `/maintenance-checklist-template/{id}` | DELETE | Remove template |
| POST | `/maintenance-checklist-template/{id}/items` | CREATE | Adiciona item ao template |
| PUT | `/maintenance-checklist-template/{id}/items/{itemId}` | UPDATE | Atualiza item |
| DELETE | `/maintenance-checklist-template/{id}/items/{itemId}` | DELETE | Remove item |

### Dashboard de Manutenção

Todos os endpoints exigem READ no módulo `"Manutenção"`. Aceitam os seguintes query params:

| Parâmetro | Tipo | Descrição |
|---|---|---|
| `clientCode` | string | Filtro por cliente |
| `assignedToId` | int | Filtro por técnico |
| `statusId` | int | Filtro por status |
| `priorityId` | int | Filtro por prioridade |
| `reasonId` | int | Filtro por motivo |
| `dateFrom` | date | Data início |
| `dateTo` | date | Data fim |
| `city` | string | Filtro por cidade |
| `state` | string | Filtro por estado |

| Método | Rota | Descrição |
|---|---|---|
| GET | `/maintenance-dashboard/summary` | Total de ordens, abertas, fechadas, vencidas |
| GET | `/maintenance-dashboard/by-status` | Contagem por status |
| GET | `/maintenance-dashboard/by-priority` | Contagem por prioridade |
| GET | `/maintenance-dashboard/by-reason` | Contagem por motivo |
| GET | `/maintenance-dashboard/by-month` | Contagem mensal (últimos 6 meses por padrão) |
| GET | `/maintenance-dashboard/by-client` | Top 10 clientes por volume |
| GET | `/maintenance-dashboard/by-requester` | Contagem por solicitante |
| GET | `/maintenance-dashboard/by-city` | Contagem por cidade |
| GET | `/maintenance-dashboard/time-entries` | Resumo de horas apontadas por tipo |
| GET | `/maintenance-dashboard/trend` | Variação percentual mês a mês |
| GET | `/maintenance-dashboard/by-technician` | Contagem e horas por técnico |

---

## Próximos módulos

- logistic
- financial
- fiscal
- commercial
- marketing
- return
- notifications
