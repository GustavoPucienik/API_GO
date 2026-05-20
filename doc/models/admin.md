# Modelos — Módulo Admin

## user

Tabela central de usuários da plataforma.

| Coluna | Tipo | Descrição |
|---|---|---|
| `user_id` | INT PK | Identificador |
| `name` | VARCHAR(255) | Primeiro nome |
| `last_name` | VARCHAR(255) | Sobrenome (opcional) |
| `email` | VARCHAR(255) UNIQUE | Email de login |
| `password` | VARCHAR(255) | Hash bcrypt da senha |
| `is_active` | TINYINT(1) | 0=inativo, 1=ativo |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Última atualização |
| `deleted_at` | DATETIME | Soft delete (NULL = ativo) |

**Regras:**
- Soft delete: registros com `deleted_at IS NOT NULL` são ignorados nas consultas.
- Senha nunca é retornada nas respostas da API.
- `is_active=0` impede login mesmo com credenciais corretas.

---

## groups

Grupos hierárquicos de usuários. Um grupo pode ter um grupo pai, formando uma árvore.

| Coluna | Tipo | Descrição |
|---|---|---|
| `group_id` | INT PK | Identificador |
| `name` | VARCHAR(255) UNIQUE | Nome do grupo |
| `parent_id` | INT FK(groups) | Grupo pai (NULL = raiz) |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Última atualização |
| `deleted_at` | DATETIME | Soft delete |

**Regras:**
- `parent_id` usa `ON DELETE SET NULL`: excluir um grupo pai não remove os filhos.
- O `data_scope=GROUP` considera todos os grupos do usuário para filtrar dados.

---

## users_groups

Relacionamento N:N entre usuários e grupos.

| Coluna | Tipo | Descrição |
|---|---|---|
| `user_group_id` | INT PK | Identificador |
| `user_id` | INT FK(user) | Usuário |
| `group_id` | INT FK(groups) | Grupo |

**Regras:**
- `UNIQUE(user_id, group_id)` — um usuário não pode estar no mesmo grupo duas vezes.
- Cascade delete: remover usuário ou grupo apaga o vínculo.

---

## access_level

Nível de acesso (perfil de permissão). Define o que um grupo pode fazer em cada módulo.

| Coluna | Tipo | Descrição |
|---|---|---|
| `access_level_id` | INT PK | Identificador |
| `name` | VARCHAR(255) UNIQUE | Nome do nível (ex: "Técnico", "Gerente") |

---

## users_access_levels

Atribuição direta de nível de acesso a um usuário (sem intermediário de grupo).

| Coluna | Tipo | Descrição |
|---|---|---|
| `user_access_level_id` | INT PK | Identificador |
| `user_id` | INT FK(user) | Usuário |
| `access_level_id` | INT FK(access_level) | Nível de acesso |

---

## groups_access_levels

Atribuição de nível de acesso a um grupo inteiro.

| Coluna | Tipo | Descrição |
|---|---|---|
| `group_access_level_id` | INT PK | Identificador |
| `group_id` | INT FK(groups) | Grupo |
| `access_level_id` | INT FK(access_level) | Nível de acesso |

---

## module

Módulos do sistema (ex: "Manutenção", "Logística"). Cada módulo tem permissões independentes.

| Coluna | Tipo | Descrição |
|---|---|---|
| `module_id` | INT PK | Identificador |
| `name` | VARCHAR(255) UNIQUE | Nome do módulo |

---

## access_levels_modules

Define o que um nível de acesso pode fazer em um módulo específico.

| Coluna | Tipo | Descrição |
|---|---|---|
| `access_level_module_id` | INT PK | Identificador |
| `access_level_id` | INT FK(access_level) | Nível de acesso |
| `module_id` | INT FK(module) | Módulo |
| `data_scope` | ENUM | `OWN` / `RESPONSIBLE` / `GROUP` / `ALL` |
| `permission` | INT | Bitmask: READ=1, CREATE=2, UPDATE=4, DELETE=8 |
| `can_be_assigned` | TINYINT(1) | Se usuários deste nível podem ser atribuídos como responsáveis |
| `created_at` | DATETIME | Criação |
| `updated_at` | DATETIME | Última atualização |
| `deleted_at` | DATETIME | Soft delete |

**Bitmask de permissão:**

| Bit | Valor | Operação |
|---|---|---|
| READ | 1 | Visualizar |
| CREATE | 2 | Criar |
| UPDATE | 4 | Editar |
| DELETE | 8 | Excluir |

Para ter CREATE e READ: `permission = 3` (1+2).

**Data scope:**

| Valor | Comportamento |
|---|---|
| `OWN` | Vê apenas registros que criou |
| `RESPONSIBLE` | Vê registros em que está como responsável |
| `GROUP` | Vê registros de usuários do mesmo grupo |
| `ALL` | Vê todos os registros |

O middleware resolve o escopo mais permissivo entre todos os `access_level`s do usuário.

---

## menu

Itens de menu exibidos no frontend, vinculados a módulos.

| Coluna | Tipo | Descrição |
|---|---|---|
| `menu_id` | INT PK | Identificador |
| `name` | VARCHAR(255) | Rótulo exibido |
| `sort_index` | INT | Ordem de exibição |
| `paths` | VARCHAR(255) | Caminho(s) de rota do frontend |
| `module_id` | INT FK(module) | Módulo ao qual o menu pertence |
| `parent_id` | INT FK(menu) | Menu pai (NULL = item de nível superior) |
| `required_scope` | ENUM | Escopo mínimo para exibir o item (`OWN`/`RESPONSIBLE`/`GROUP`/`ALL`) |

**Regras:**
- Sem soft delete. Menus são excluídos fisicamente.
- `required_scope` permite esconder itens de menu para usuários com escopo restrito.
