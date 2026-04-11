# Memory Integrity Validation Matrix

## Objetivo

Garantir que a memória central do Brain Sentry:

1. entra corretamente no core;
2. é persistida sem perda de campos críticos;
3. respeita isolamento por tenant;
4. aparece corretamente nas telas do admin;
5. sai corretamente das listagens quando expirada, supersedida ou deletada.

## Fluxo Real da Memória

1. `POST /v1/memories` entra por [memory.go](/Users/edsonmartins/desenvolvimento/brainsentry.io/brain-sentry-go/internal/handler/memory.go)
2. o core aplica regras em [memory.go](/Users/edsonmartins/desenvolvimento/brainsentry.io/brain-sentry-go/internal/service/memory.go)
3. a persistência acontece em [memory.go](/Users/edsonmartins/desenvolvimento/brainsentry.io/brain-sentry-go/internal/repository/postgres/memory.go)
4. o admin consome essas leituras via:
   - [MemoryAdminPage.tsx](/Users/edsonmartins/desenvolvimento/brainsentry.io/brain-sentry-frontend/src/pages/MemoryAdminPage.tsx)
   - [DashboardPage.tsx](/Users/edsonmartins/desenvolvimento/brainsentry.io/brain-sentry-frontend/src/pages/DashboardPage.tsx)
   - [SearchPage.tsx](/Users/edsonmartins/desenvolvimento/brainsentry.io/brain-sentry-frontend/src/pages/SearchPage.tsx)
   - [TimelinePage.tsx](/Users/edsonmartins/desenvolvimento/brainsentry.io/brain-sentry-frontend/src/pages/TimelinePage.tsx)
   - [RelationshipsPage.tsx](/Users/edsonmartins/desenvolvimento/brainsentry.io/brain-sentry-frontend/src/pages/RelationshipsPage.tsx)
   - [AuditPage.tsx](/Users/edsonmartins/desenvolvimento/brainsentry.io/brain-sentry-frontend/src/pages/AuditPage.tsx)

## Invariantes de Integridade

| Invariante | Automação | Conferência visual no admin | Observação |
| --- | --- | --- | --- |
| `content`, `summary`, `category`, `importance` persistem corretamente | Sim, integração do repositório | Sim, Memórias / Dashboard / Busca / Timeline | Base mínima |
| `tags` persistem e retornam na leitura | Sim, integração do repositório | Parcial, Memórias / Timeline / Relationships | Não aparece em todas as telas |
| `metadata` persiste sem perda | Sim, integração do repositório | Não | Precisa validação por API/DB |
| `tenant_id` isola leitura e listagem | Sim, integração do repositório | Não | Admin não expõe tenant por memória |
| `deleted_at` remove a memória das listagens | Sim, integração do repositório | Sim, a memória some do admin | Campo em si não é visível |
| `sim_hash` persiste para deduplicação | Sim, integração do repositório | Não | Precisa API/DB |
| `valid_from` / `valid_to` governam memória ativa | Sim, integração do repositório | Parcial, efeito é visível; campo não | Válido para expiração |
| `superseded_by` exclui memória supersedida das ativas | Sim, integração do repositório | Parcial, efeito é visível; campo não | Importante para integridade temporal |
| `version` cresce em update/rollback | Parcial | Parcial, via modal de versões | Falta prova ponta a ponta com atualização real |
| `access_count` / `helpful_count` / `injection_count` refletem uso | Parcial | Sim, Dashboard / Memórias / Analytics | Falta cenário real sem mock |
| busca encontra a memória correta | Sim, Playwright com contrato de UI | Sim, Busca | Hoje a suíte é determinística por mocks |
| criação aparece em telas derivadas | Sim, Playwright com contrato de UI | Sim | Ainda não é backend real |

## O Que Já Está Coberto

### Core / Persistência

- CRUD básico de memória
- full-text search
- feedback
- soft delete
- preservação de campos críticos
- isolamento por tenant
- expiração temporal
- exclusão de memórias supersedidas da lista ativa

### Admin / Tela

- autenticação
- navegação
- dashboard
- memórias
- busca
- timeline
- relacionamentos
- auditoria
- usuários
- tenants
- configurações
- analytics
- perfil
- playground
- conectores
- notas
- tarefas

## O Que Ainda Não Dá Para Garantir Só Pela Tela

- valor bruto de `metadata`
- valor bruto de `sim_hash`
- `deleted_at`
- `tenant_id` por memória
- `superseded_by`
- `valid_from` e `valid_to`
- embedding gerado

Para esses campos, a validação correta precisa ser por:

- teste de integração do repositório;
- teste de API;
- ou inspeção direta no banco.

## Estratégia Recomendada

### Camada 1: Core

- integração PostgreSQL para provar persistência, isolamento e ciclo de vida
- testes de serviço para deduplicação, versionamento e scoring

### Camada 2: API

- smoke tests reais contra `/v1/memories`, `/v1/memories/search`, `/v1/memories/{id}/versions`
- validação de payload completo antes/depois de update/delete

### Camada 3: Admin

- Playwright para provar que o estado retornado pela API aparece corretamente em tela
- execução `headed` para inspeção visual quando necessário

## Checklist de Validação Manual com Backend Real

1. Criar uma memória pelo admin.
2. Confirmar a presença dela em `Memórias`.
3. Confirmar a presença dela em `Busca`.
4. Confirmar a presença dela em `Timeline`.
5. Confirmar métricas no `Dashboard`.
6. Editar a memória e abrir `Histórico de Versões`.
7. Buscar por tag/categoria/importância.
8. Deletar a memória e confirmar que some das listagens.
9. Validar o registro correspondente em `Auditoria`.
10. Conferir por API ou DB os campos não visíveis em tela.
