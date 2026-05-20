# Modelos — Módulo Manutenção

## Lookups (tabelas de configuração)

### maintenance_order_status

Status de uma ordem de manutenção. Controla o ciclo de vida da ordem.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_status_id` | INT PK | Identificador |
| `name` | VARCHAR(100) UNIQUE | Nome do status |
| `background_color` | VARCHAR(20) | Cor de fundo (hex ou CSS) |
| `border_color` | VARCHAR(20) | Cor da borda |
| `text_color` | VARCHAR(20) | Cor do texto |
| `sort_index` | INT | Ordem de exibição |
| `is_closed` | TINYINT(1) | 1 = status de encerramento |
| `is_default` | TINYINT(1) | 1 = status atribuído automaticamente na criação |

**Regras:**
- Sem soft delete (hard DELETE).
- `is_default=1`: ao criar uma ordem sem informar status, o sistema busca o primeiro status com esse flag.
- `is_closed=1`: ao mover para este status, `closed_at` da ordem é preenchido automaticamente. Ao sair, `closed_at` é limpo.

---

### maintenance_order_priorities

Prioridade de uma ordem (ex: Baixa, Média, Alta, Urgente).

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_priorities_id` | INT PK | Identificador |
| `name` | VARCHAR(100) UNIQUE | Nome |
| `background_color` | VARCHAR(20) | Cor de fundo |
| `border_color` | VARCHAR(20) | Cor da borda |
| `text_color` | VARCHAR(20) | Cor do texto |
| `sort_index` | INT | Ordem de exibição |

**Regras:** Sem soft delete (hard DELETE).

---

### maintenance_order_reason

Motivo/categoria da ordem de manutenção.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_reason_id` | INT PK | Identificador |
| `name` | VARCHAR(100) UNIQUE | Nome do motivo |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

**Relação:** Templates de checklist podem ser vinculados a um motivo, aparecendo automaticamente para ordens daquele motivo.

---

### maintenance_order_note_template

Modelos de texto pré-definidos para agilizar o preenchimento de descrições/resoluções.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_note_template_id` | INT PK | Identificador |
| `name` | VARCHAR(255) UNIQUE | Título do modelo |
| `description` | TEXT | Conteúdo do modelo |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

---

### maintenance_order_time_entry_type

Tipos de apontamento de tempo (ex: Deslocamento, Atendimento, Aguardando Peça).

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_time_entry_type_id` | INT PK | Identificador |
| `name` | VARCHAR(100) UNIQUE | Nome do tipo |
| `color` | VARCHAR(20) | Cor de destaque (opcional) |

**Regras:** Sem soft delete (hard DELETE).

---

## Ordem de Manutenção

### maintenance_order

Tabela principal. Cada linha representa uma ordem de serviço.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_id` | INT PK | Identificador |
| `requester_id` | INT FK(user) | ID do usuário que criou |
| `requester_name` | VARCHAR(255) | Snapshot do nome do solicitante na criação |
| `salesperson_name` | VARCHAR(255) | Nome do vendedor do cliente (snapshot) |
| `client_code` | VARCHAR(50) | Código do cliente no SAP |
| `client_name` | VARCHAR(255) | Razão social (snapshot) |
| `client_alias_name` | VARCHAR(255) | Nome fantasia (snapshot) |
| `contact_type` | VARCHAR(100) | Tipo do contato (opcional) |
| `contact_name` | VARCHAR(255) | Nome do contato |
| `contact_phone` | VARCHAR(50) | Telefone do contato |
| `contact_mobile` | VARCHAR(50) | Celular do contato |
| `contact_email` | VARCHAR(255) | Email do contato |
| `address_name` | VARCHAR(100) | Nome do endereço |
| `address_type` | VARCHAR(50) | Tipo de endereço |
| `type_of_address` | VARCHAR(100) | Subtipo de endereço |
| `street` | VARCHAR(255) | Logradouro |
| `street_no` | VARCHAR(50) | Número |
| `block` | VARCHAR(100) | Bairro |
| `city` | VARCHAR(100) | Cidade |
| `county` | VARCHAR(100) | Município |
| `state` | VARCHAR(100) | Estado |
| `country` | VARCHAR(100) | País |
| `zip_code` | VARCHAR(20) | CEP |
| `building_floor_room` | VARCHAR(255) | Complemento |
| `status_id` | INT FK(status) | Status atual |
| `priority_id` | INT FK(priorities) | Prioridade |
| `reason_id` | INT FK(reason) | Motivo |
| `description` | TEXT | Descrição do problema |
| `captured_at` | DATETIME | Data/hora em que o problema foi capturado |
| `resolution` | TEXT | Texto de resolução |
| `motivo_resolution` | VARCHAR(255) | Motivo da resolução |
| `client_signature` | LONGTEXT | Assinatura do cliente em base64 |
| `closed_at` | DATETIME | Preenchido automaticamente ao atingir status `is_closed=1` |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

**Regras:**
- Dados de cliente e solicitante são copiados como snapshot na criação para preservar o histórico mesmo que o cadastro mude.
- `closed_at` é gerenciado pela API: setado ao mudar para status fechado, limpo ao reabrir.
- A filtragem por escopo (`OWN`/`RESPONSIBLE`/`GROUP`/`ALL`) é aplicada na listagem paginada.

---

### maintenance_order_items

Itens/peças solicitados na ordem (material inicial).

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_item_id` | INT PK | Identificador |
| `maintenance_order_id` | INT FK(order) | Ordem |
| `name` | VARCHAR(255) | Nome do item (snapshot do SAP) |
| `item_code` | VARCHAR(50) | Código no SAP |
| `quantity` | INT | Quantidade |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

**Regras:** O código é validado contra a tabela `sap_items` antes de salvar. O nome é copiado como snapshot.

---

### maintenance_order_resolution_items

Itens/peças utilizados na resolução da ordem.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_resolution_item_id` | INT PK | Identificador |
| `maintenance_order_id` | INT FK(order) | Ordem |
| `name` | VARCHAR(255) | Nome do item (snapshot) |
| `item_code` | VARCHAR(50) | Código no SAP |
| `quantity` | INT | Quantidade |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

---

### maintenance_order_time_entry

Apontamentos de tempo realizados em uma ordem.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_time_entry_id` | INT PK | Identificador |
| `maintenance_order_id` | INT FK(order) | Ordem |
| `type_id` | INT FK(time_entry_type) | Tipo de apontamento |
| `started_at` | TIMESTAMP | Início |
| `ended_at` | TIMESTAMP | Fim (NULL = em aberto) |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

---

## Agendamento

### maintenance_order_schedule_date

Datas agendadas para atendimento de uma ordem. Uma ordem pode ter múltiplas datas.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_schedule_date_id` | INT PK | Identificador |
| `maintenance_order_id` | INT FK(order) | Ordem |
| `date` | DATETIME | Data/hora do agendamento |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

---

### maintenance_order_schedule_date_technician

Técnicos atribuídos a uma data de agendamento.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_schedule_date_technician_id` | INT PK | Identificador |
| `maintenance_order_schedule_date_id` | INT FK(schedule_date) | Data de agendamento |
| `technician_id` | INT FK(user) | ID do técnico |
| `technician_name` | VARCHAR(255) | Nome snapshot do técnico |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

**Regras:**
- Ao atribuir (`PATCH /maintenance/{id}/assign`), a API verifica se o técnico já está em outra ordem na mesma data.
- O nome é copiado como snapshot.
- Cascade delete: remover a data de agendamento apaga os técnicos vinculados.

---

## Anexos

### maintenance_order_attachment

Arquivos enviados para uma ordem.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_attachment_id` | INT PK | Identificador |
| `maintenance_order_id` | INT FK(order) | Ordem |
| `file_name` | VARCHAR(255) | Nome original do arquivo |
| `file_path` | VARCHAR(500) | Caminho físico salvo (ex: `uploads/orders/1716123456789_foto.jpg`) |
| `file_size` | INT | Tamanho em bytes |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

**Regras:**
- O arquivo físico é salvo em `uploads/orders/{timestamp_ns}_{nome_original}`.
- Ao deletar, o registro recebe soft delete E o arquivo físico é removido do disco.

---

## Checklists

### maintenance_order_checklist_template

Template (modelo) de checklist que pode ser aplicado a uma ordem.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_checklist_template_id` | INT PK | Identificador |
| `name` | VARCHAR(255) | Nome do template |
| `reason_id` | INT FK(reason) | Motivo associado (NULL = genérico) |
| `active` | TINYINT(1) | 1 = ativo e disponível para uso |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

**Regras:** Templates vinculados a um `reason_id` aparecem sugeridos para ordens daquele motivo.

---

### maintenance_order_checklist_item

Itens (perguntas/campos) de um template de checklist.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_checklist_item_id` | INT PK | Identificador |
| `template_id` | INT FK(template) | Template ao qual pertence |
| `section` | VARCHAR(255) | Seção/grupo visual (opcional) |
| `label` | TEXT | Texto da pergunta ou instrução |
| `type` | ENUM | Tipo do campo (ver abaixo) |
| `options` | TEXT | JSON com opções para `single_choice` / `multi_choice` |
| `metadata` | TEXT | Metadados adicionais (JSON livre) |
| `required` | TINYINT(1) | 1 = resposta obrigatória |
| `sort_order` | INT | Ordem de exibição |
| `parent_item_id` | INT FK(self) | Item pai (para itens condicionais) |
| `show_when_values` | TEXT | JSON com valores do pai que ativam este item |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

**Tipos de item:**

| Tipo | Descrição |
|---|---|
| `single_choice` | Seleção única (rádio) — opções em `options` |
| `multi_choice` | Seleção múltipla (checkbox) — opções em `options` |
| `text` | Campo de texto livre |
| `table_row` | Linha de tabela estruturada (colunas em `metadata`) |
| `sap_item` | Seleção de item do catálogo SAP |

---

### maintenance_order_checklist

Instância de um checklist preenchido para uma ordem específica.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_checklist_id` | INT PK | Identificador |
| `maintenance_order_id` | INT FK(order) | Ordem |
| `template_id` | INT FK(template) | Template utilizado |
| `completed_at` | DATETIME | Momento da conclusão (NULL = rascunho) |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

---

### maintenance_order_checklist_answer

Respostas de cada item do checklist preenchido.

| Coluna | Tipo | Descrição |
|---|---|---|
| `maintenance_order_checklist_answer_id` | INT PK | Identificador |
| `checklist_id` | INT FK(checklist) | Checklist instanciado |
| `item_id` | INT FK(checklist_item) | Item respondido |
| `value` | TEXT | Resposta (texto, JSON de seleções, código SAP, etc.) |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Atualização |
| `deleted_at` | DATETIME | Soft delete |

---

## Diagrama de dependências

```
user ──────────────────────────────────────────────────┐
  └── users_groups ── groups ── groups_access_levels ──┤
  └── users_access_levels ────────────────────────────── access_level
                                                              └── access_levels_modules ── module ── menu

maintenance_order_status ──────────────────────────────┐
maintenance_order_priorities ──────────────────────────┤
maintenance_order_reason ──────────────────────────────┴── maintenance_order
                                                              ├── maintenance_order_items
                                                              ├── maintenance_order_resolution_items
                                                              ├── maintenance_order_time_entry ── maintenance_order_time_entry_type
                                                              ├── maintenance_order_schedule_date
                                                              │     └── maintenance_order_schedule_date_technician
                                                              ├── maintenance_order_attachment
                                                              └── maintenance_order_checklist
                                                                    ├── maintenance_order_checklist_template
                                                                    │     └── maintenance_order_checklist_item
                                                                    └── maintenance_order_checklist_answer
```
