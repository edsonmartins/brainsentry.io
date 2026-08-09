# Auditoria final e cobertura de testes

Data da repetição: 9 de agosto de 2026.

## Conclusão executiva

O núcleo do produto está coerente com a proposta descrita em
[`PRODUCT_CAPABILITIES.md`](PRODUCT_CAPABILITIES.md): memórias são persistidas
com isolamento por tenant, histórico bi-temporal, busca híbrida, projeção para
grafo, recuperação para prompts, expiração e trilha de auditoria. A repetição da
auditoria encontrou e corrigiu regressões que a primeira passagem não havia
encerrado: campos opcionais quebrando histórico, motivo de versão perdido,
eventos de outbox presos, índice/vetores incompatíveis com FalkorDB e tarefas
LLM agendadas sem provedor. Uma revisão posterior também detectou que o
backfill original da migration `000016` usava nomes de colunas PostgreSQL no
JSON histórico; instalações novas foram corrigidas e ambientes que já haviam
aplicado a migration recebem o reparo idempotente `000017`.

Isso não equivale a afirmar que toda capacidade periférica está exaustivamente
testada. O núcleo tem evidência unitária, de integração e E2E real; capacidades
dependentes de um provedor LLM externo ainda precisam de uma suíte contratual
com credencial de homologação.

## Princípio de armazenamento validado

- PostgreSQL é o sistema de registro canônico.
- `memory_history` conserva a verdade por versão e sustenta consultas `as of`.
- `projection_outbox` é gravada na mesma transação da mutação canônica.
- FalkorDB é uma projeção reconstruível, isolada por `tenantId`, e recebe
  embeddings como `vecf32` com índice criado no boot.
- Redis é aceleração e coordenação; sua ausência não pode alterar a verdade.
- Exclusão limpa a memória canônica e produz uma projeção de remoção do grafo.
- Conteúdo externo usado em prompts passa por sanitização e enquadramento de
  memória não confiável.

## Matriz capacidade → evidência

| Capacidade | Evidência principal | Nível | Resultado |
|---|---|---|---|
| CRUD, campos derivados e idempotência | service/repository + `real-memory-integrity.spec.ts` | unitário + PostgreSQL real + E2E real | aprovado |
| Histórico/versionamento bi-temporal | integração PostgreSQL e E2E de versões/motivo | integração + E2E real | aprovado |
| Multi-tenancy | repository, store fail-closed e queries de grafo | unitário + integração | aprovado |
| Outbox e projeção no grafo | claim/ack/retry/reclaim e worker simulado | integração + unitário | aprovado |
| Busca e interceptação | service/repository e memórias ativas/expiradas | unitário + E2E real | aprovado |
| Busca vetorial FalkorDB | boot/consulta com FalkorDB real e testes de fallback | integração real + unitário | aprovado |
| Auto-forget e consolidação | service e lifecycle real | unitário + E2E real | aprovado, conservador |
| Segurança de prompt | pacote `security` e pipelines LLM | unitário | aprovado |
| MCP JSON-RPC/SSE | suíte `internal/mcp` | unitário/handler | aprovado, sem cliente externo E2E |
| Admin React | 74 cenários com API simulada | E2E simulado | aprovado |
| Admin + backend real | 3 cenários de integridade/lifecycle | E2E real | aprovado |
| Capacidades LLM | providers simulados, sanitização e `llm_smoke` opt-in | contrato + gate real | gate pronto; execução aguarda secret de homologação |
| CLI | suíte de comandos | unitário | aprovado parcialmente |
| TUI | navegação, atalhos, confirmação, charts, status e toast | unitário comportamental | aprovado parcialmente |

## Medições reproduzidas

- Go, suíte curta completa: **41,0%** das statements.
- `internal/service`: **52,1%**.
- `internal/security`: **95,2%**.
- `internal/mcp`: **57,0%**.
- `internal/handler`: **20,4%** (antes: 15,7%).
- `internal/repository/graph`: **70,6%** (antes: 34,7%).
- PostgreSQL com Testcontainers e migrations reais: **27,7%** do pacote.
- TUI principal: **27,2%**; componentes TUI: **79,0%**.
- Frontend Vitest/V8: **1,03% statements**, **26,51% branches**,
  **11,5% functions** e **1,03% lines**, com baseline bloqueando regressão.
- Frontend Playwright simulado: **74/74** no Chromium e **74/74** no Firefox.
- Frontend/backend/PostgreSQL/FalkorDB reais: **3/3** no Chromium e **3/3** no Firefox.
- Build Go e build de produção React: aprovados.
- ESLint: zero erros e **310 warnings** legados após o code splitting.

Cobertura percentual não é usada isoladamente como prova de produto. O pacote
PostgreSQL, por exemplo, tem percentual modesto, mas os testes de integração
atingem as invariantes críticas de mutação atômica, histórico e outbox. Já o
frontend tem boa cobertura de fluxos E2E e agora publica statements de
componentes, mas o baseline global inicial ainda é baixo e deve crescer por
fluxos críticos.

## Débitos e limites remanescentes

1. Executar o gate `go test -tags=llm_smoke ./internal/service -run
   TestLiveLLMProvider` com secret de homologação. O gate aceita OpenRouter,
   Anthropic ou Gemini e falha explicitamente se nenhuma credencial existir.
   Depois, ampliar o smoke para compressão, triplets, eventos e coreference.
2. Continuar elevando handlers (agora 20,4%) e PostgreSQL (27,7% em integração),
   priorizando autorização, paginação e respostas de erro. O grafo chegou a
   70,6% e ganhou testes de isolamento tenant nas escritas de entidades.
3. Ampliar TUI para os modelos de cada view; navegação e componentes já possuem
   cobertura comportamental, mas `cmd/tui/views` permanece sem instrumentação.
4. Elevar o baseline global do frontend, hoje 1,03% de statements. A publicação
   e os thresholds já estão no `pnpm test:coverage`; os E2E continuam como gate
   separado porque não equivalem à cobertura unitária de statements.
5. Tratar os 310 warnings de lint, priorizando acessibilidade e dependências de
   hooks. O lint agora é executável, mas o legado foi inicialmente classificado
   como warning para permitir redução incremental.
6. Manter o code splitting por rota e o orçamento automático do build. Após a
   correção, o entrypoint caiu de aproximadamente 2,34 MB/712 KB gzip para
   638 KB/206 KB gzip; o build rejeita regressões acima de 250 KB gzip no
   entrypoint ou 190 KB gzip em chunks assíncronos.
7. Adicionar WebKit à homologação. Chromium e Firefox estão aprovados tanto na
   suíte simulada quanto nos três fluxos com backend e bancos reais.

## Gates recomendados

Em cada PR: build Go, suíte Go curta, build/lint frontend e Playwright simulado.
Em merge/release: integração PostgreSQL, E2E real com PostgreSQL/FalkorDB e um
smoke LLM protegido por secret. Percentuais devem subir por pacote crítico; uma
meta global única favoreceria testes periféricos sem proteger as invariantes do
produto.
