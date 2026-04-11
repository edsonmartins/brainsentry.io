# Brain Sentry Core Product Test Strategy

## Produto Core

O core do Brain Sentry nao e o admin web em si. O admin e uma superficie de observabilidade e operacao. O produto central e o sistema de memoria para agentes de IA:

1. Receber eventos, prompts e fatos.
2. Transformar conteudo bruto em memorias persistentes e multi-tenant.
3. Classificar, deduplicar, versionar e preservar integridade.
4. Recuperar memorias relevantes por busca lexical, vetorial e grafo.
5. Re-ranquear por score hibrido, decaimento temporal, importancia, tags e peso emocional.
6. Injetar contexto no prompt de forma autonoma, com budget de tokens e mascaramento de PII.
7. Aprender com sessoes, reconciliar fatos, consolidar memorias e esquecer/supersedar informacao obsoleta.
8. Expor tudo isso via API/MCP/admin com auditabilidade.

## Rotinas Criticas

### 1. Memory Formation and Storage

Arquivos principais:

- `brain-sentry-go/internal/service/memory.go`
- `brain-sentry-go/internal/repository/postgres/memory.go`
- `brain-sentry-go/internal/repository/postgres/version.go`
- `brain-sentry-go/internal/domain/memory.go`

Invariantes:

- `content`, `summary`, `category`, `importance`, `memoryType`, `metadata`, `tags`, `tenantId`, `sourceType`, `sourceReference`, `createdBy`, `emotionalWeight`, `validFrom`, `validTo`, `simHash`, `decayRate` devem persistir sem perda.
- `memoryType` deve ser classificado quando ausente.
- `decayRate` deve derivar de `memoryType`.
- `emotionalWeight` deve ser limitado entre -1 e 1.
- `simHash` deve ser gerado para dedup/supersession.
- near-duplicate com Hamming <= 3 nao deve criar duplicata; deve retornar memoria existente e aumentar acesso.
- cada criacao/edicao deve produzir historico de versao.
- soft delete deve remover a memoria de list/search/get sem destruir historico.
- isolamento por tenant deve impedir vazamento entre tenants.

Testes prioritarios:

- Unitarios de `MemoryService` com repos fake para dedup, defaults, classificacao e versionamento.
- Integracao Postgres para campos brutos, tags, soft delete, tenant isolation, versions e full-text.
- E2E real para criar/editar/versionar/deletar e conferir no admin.

### 2. Retrieval and Ranking

Arquivos principais:

- `brain-sentry-go/internal/service/scoring.go`
- `brain-sentry-go/internal/service/rrf_scoring.go`
- `brain-sentry-go/internal/service/retrieval_planner.go`
- `brain-sentry-go/internal/service/reranker.go`
- `brain-sentry-go/internal/service/memory.go`

Invariantes:

- busca deve excluir memorias expiradas e supersedadas.
- score hibrido deve respeitar similaridade, overlap lexical, proximidade de grafo, recencia, tags, importancia, decaimento e peso emocional.
- RRF deve unir streams sem duplicar memoria.
- diversidade por sessao deve limitar resultados dominantes.
- planner deve cair para fallback quando LLM ou grafo nao estiver disponivel.
- resultado final deve ser deterministico quando entradas sao deterministicas.

Testes prioritarios:

- Unitarios de scoring/RRF ja existem, manter como base.
- Adicionar teste de fluxo `SearchMemories` com repo fake: resultados vetoriais + textuais, dedup entre streams, exclusao por `validTo` e `supersededBy`, ordenacao final por `RelevanceScore`.
- Integracao Postgres para full-text search retornando somente tenant atual e memorias ativas.
- Benchmark de qualidade com dataset pequeno e expectativas fixas para recall/MRR/NDCG.

### 3. Autonomous Context Injection

Arquivos principais:

- `brain-sentry-go/internal/service/interception.go`
- `brain-sentry-go/internal/handler/interception.go`
- `brain-sentry-go/internal/mcp/tools.go`

Invariantes:

- prompt curto deve retornar sem enriquecimento.
- quick check deve evitar busca quando nao ha padroes relevantes, exceto `forceDeepAnalysis`.
- deep analysis deve respeitar threshold de confianca.
- memoria expirada/supersedada nao pode entrar no contexto.
- contexto deve respeitar budget de tokens.
- PII deve ser mascarado antes de compor `enhancedPrompt`.
- hindsight notes devem entrar somente em prompts com sinais de erro.
- `MemoriesUsed`, `TokensInjected`, `LLMCalls`, `LatencyMs` e contadores de injecao devem ser consistentes.

Testes prioritarios:

- Unitarios de budget, PII, filtros e quick/deep gating.
- Teste de servico com repos fake para fluxo completo: prompt relevante -> busca -> filtro -> contexto -> contadores.
- E2E real API `/v1/intercept`: criar memorias relevantes, chamar intercept, validar `enhanced=true`, `ContextInjected` contem memoria correta e nao contem memoria expirada/supersedada.
- Teste MCP `intercept_prompt` validando o mesmo contrato pelo protocolo.

### 4. Learning Lifecycle

Arquivos principais:

- `brain-sentry-go/internal/service/reconciliation.go`
- `brain-sentry-go/internal/service/consolidation.go`
- `brain-sentry-go/internal/service/semantic_memory.go`
- `brain-sentry-go/internal/service/reflection.go`
- `brain-sentry-go/internal/service/auto_forget.go`
- `brain-sentry-go/internal/service/cross_session.go`

Invariantes:

- reconciliation deve ADD/UPDATE/DELETE/NONE sem quebrar versionamento nem tenant.
- DELETE por reconciliation deve supersedar, nao destruir sem rastro.
- consolidation deve preservar informacao unica, tags, maior importancia e contadores.
- semantic consolidation deve gerar memorias `SEMANTIC` ou `PROCEDURAL` com metadata rastreavel.
- auto-forget em dry-run nao pode alterar dados.
- auto-forget real deve respeitar `MaxDeletesPerRun`.
- cross-session deve aplicar redaction e gerar memorias episodicas com proveniencia.

Testes prioritarios:

- Unitarios com LLM fake para reconciliation ADD/UPDATE/DELETE/NONE.
- Unitarios com LLM fake para consolidation merge/compress.
- Integracao Postgres para auto-forget TTL, low-value e supersession.
- E2E API real para `/v1/reconcile`, `/v1/consolidate`, `/v1/semantic/consolidate`, `/v1/auto-forget` com dataset controlado.

### 5. Graph and Associative Memory

Arquivos principais:

- `brain-sentry-go/internal/service/spreading_activation.go`
- `brain-sentry-go/internal/service/entity_graph.go`
- `brain-sentry-go/internal/service/nl_cypher.go`
- `brain-sentry-go/internal/service/louvain.go`
- `brain-sentry-go/internal/repository/graph`

Invariantes:

- ativacao deve decair por hop e respeitar threshold.
- ciclos no grafo nao devem causar loop infinito.
- relacionamento mais forte deve vencer ativacao menor existente.
- NL to Cypher deve escapar entradas e falhar seguro.
- quando FalkorDB estiver indisponivel, o produto deve degradar sem quebrar APIs principais.

Testes prioritarios:

- Unitarios de propagacao com graph fake.
- Integracao opcional com FalkorDB para grafo pequeno A-B-C validando ativacao e consulta NL controlada.
- E2E API real de `/v1/memories/activate` quando FalkorDB estiver ativo.

## Camadas de Teste Recomendadas

### Tier 1: Algoritmos Puros

Executar sempre em CI:

- `go test ./internal/service`
- foco em scoring, decay, simhash, classifier, PII, compression, RRF, planner, learning.

### Tier 2: Servicos com Fakes

Executar sempre em CI:

- repositorios fake em memoria para `MemoryService`, `InterceptionService`, `ReconciliationService`, `ConsolidationService`.
- LLM fake com respostas deterministicas.
- graph fake para ativacao e retrieval.

### Tier 3: Integracao Postgres

Executar em CI com Postgres:

- `go test -tags=integration ./internal/repository/postgres`
- validar schema, migrations, tenant isolation, soft delete, versions, tags, full-text e campos brutos.

### Tier 4: API Real

Executar com backend real:

- autenticar via `/v1/auth/demo`.
- criar dataset via API.
- exercitar `/v1/memories`, `/v1/memories/search`, `/v1/intercept`, `/v1/reconcile`, `/v1/consolidate`, `/v1/auto-forget`.
- validar bruto via API, nao apenas status HTTP.

### Tier 5: Admin Web

Executar com Playwright:

- usar admin somente para validar o que precisa ser observado visualmente: memoria criada, busca, historico, delete, dashboards e sinais de operacao.
- nao usar a tela como fonte primaria de verdade para metadata, simHash, supersededBy, validFrom/validTo ou tenantId; esses campos devem ser conferidos por API/DB.

## Proxima Sequencia de Implementacao

1. Criar testes de servico com fakes para `MemoryService.SearchMemories` cobrindo filtro de expiradas/supersedadas, dedup entre vector/text e ordenacao final. Concluido na camada de servico: o predicado compartilhado de memoria inativa foi coberto em unitario e aplicado nos fluxos vetoriais/textuais; `MemoryService` agora aceita repositorio, graph e embedding via interfaces internas e cobre fallback textual com exclusao de expiradas/supersedadas, ordenacao por score hibrido e dedup entre vector/text com graph fake.
2. Criar E2E API real para `/v1/intercept` e learning lifecycle, reutilizando o setup de `real-memory-integrity.spec.ts`. Iniciado: a suite real agora cria memorias ativa/expirada/minor, chama `/v1/intercept` e valida contexto, `MemoriesUsed` e `InjectionCount`; tambem cobre `/v1/auto-forget?dryRun=true` e `/v1/semantic/consolidate` abaixo do minimo, validando contrato e ausencia de mutacao em fixture dry-run.
3. Criar testes com LLM fake para `ReconciliationService` nos quatro caminhos ADD/UPDATE/DELETE/NONE. Concluido na camada de decisao: `ReconciliationService` agora aceita `LLMProvider`, mantendo compatibilidade com `OpenRouterService`, e cobre ADD via `ReconcileFacts` sem repositorio e UPDATE/DELETE/NONE via `decideAction`.
4. Criar testes de semantic/consolidation/cross-session e integracao Postgres para `AutoForgetService` com dry-run e real-run. Concluido nas camadas de servico e Postgres: `SemanticMemoryService` agora aceita repositorio via interface, varre a primeira pagina correta e cobre extracao com LLM fake, proveniencia, metadata e persistencia como memorias `SEMANTIC`/`PROCEDURAL` com Postgres real; `ConsolidationService` agora aceita `LLMProvider` e repositorio via interface, cobre merge de memorias similares preservando informacao unica, tags, maior importancia, contadores e versionamento, compressao de memoria longa isolada e merge com update/delete reais no Postgres; `CrossSessionService` cobre criacao de memoria `EPISODIC` com redaction de PII, tags e proveniencia de sessao no Postgres real; `AutoForgetService` agora aceita repositorio via interface, varre a primeira pagina correta, tolera auditoria ausente e cobre dry-run, TTL real, supersession por duplicidade, low-value e limite `MaxDeletesPerRun` com repos fake e com integracao Postgres real para dry-run e real-run. As integracoes tambem corrigiram persistencia de contadores no `MemoryRepository.Update` e carregamento de tags no `MemoryRepository.FindAll`.
5. Criar suite opcional com FalkorDB para ativacao associativa e GraphRAG. Parcialmente coberto na camada de servico: `SpreadingActivationService` agora aceita provider de vizinhos via interface e cobre propagacao com grafo fake, ciclos, decaimento por hop, threshold minimo e escolha do caminho mais forte; `EntityGraphService` agora aceita extrator e repositorio via interfaces internas e cobre armazenamento de entidades/relacionamentos com fakes, incluindo no-op seguro quando OpenRouter/FalkorDB nao estao disponiveis. Ainda falta integracao opcional com FalkorDB.

## Status Atual Observado

Comando executado:

```bash
go test ./internal/service ./internal/handler ./internal/repository/postgres -cover
```

Resultado:

- `internal/service`: 44.9% statements.
- `internal/handler`: 19.9% statements.
- `internal/repository/postgres`: 1.4% statements sem tag de integracao.

Leitura:

- Existem muitos testes unitarios para algoritmos pequenos.
- A cobertura mais fraca esta em fluxos integrados e persistencia.
- O foco deve sair de “cobrir todas as telas do admin” e ir para “provar invariantes do ciclo de vida da memoria”.
