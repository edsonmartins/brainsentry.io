# Brain Sentry — capacidades e funcionamento do produto

> Referência vigente sobre o problema que o produto resolve, seu modelo de memória e as capacidades observadas no backend Go. Documentos anteriores que descrevem Java/Spring, Next.js ou FalkorDB como fonte primária são históricos.

## O que é

Brain Sentry é uma camada de memória persistente para fleets de agentes de IA. Seu objetivo é preservar conhecimento entre sessões e devolver contexto relevante quando um agente precisa tomar uma decisão, continuar um trabalho ou evitar repetir um erro.

O produto opera ao redor do modelo de IA:

1. recebe conteúdo de humanos, agentes, sessões, documentos e integrações;
2. transforma esse conteúdo em memórias estruturadas e multi-tenant;
3. mantém validade, proveniência, versões, feedback e relacionamentos;
4. recupera candidatos por identidade, texto, vetor, tempo e grafo;
5. classifica e limita o contexto antes de entregá-lo ao agente;
6. corrige, consolida, supersede ou expira conhecimento ao longo do tempo.

## Problemas que resolve

- Continuidade entre sessões sem depender apenas da janela de contexto do modelo.
- Consistência de decisões, políticas, padrões e preferências.
- Conhecimento compartilhado entre agentes autorizados do mesmo tenant.
- Recuperação de incidentes e soluções anteriores.
- Redução do volume de histórico enviado ao LLM.
- Rastreabilidade da origem, validade, confiança e correções de cada memória.

## Por que não é apenas RAG

Um RAG tradicional indexa e consulta documentos. Brain Sentry também possui escrita e ciclo de vida:

- cria e classifica memórias;
- atualiza e registra versões;
- recebe feedback de utilidade;
- consolida observações em conhecimento;
- detecta duplicidade e conflitos;
- encerra a validade de fatos;
- aprende entre sessões;
- expõe proveniência e confiança.

Retrieval é uma parte do sistema de memória, não o produto inteiro.

## Modelo de memória

Cada memória pode conter conteúdo, resumo, categoria, importância, tipo cognitivo, tags, metadata, origem, tenant, autor, peso emocional, SimHash, embedding derivado, validade temporal, proveniência, status de validação e contadores de acesso, injeção e feedback.

### Tipos cognitivos

| Tipo | Papel | Exemplo |
|---|---|---|
| Semântica | Conhecimento factual ou conceitual | “O serviço de pagamentos usa idempotency keys.” |
| Episódica | Algo ocorrido em uma sessão ou instante | “O deploy falhou por timeout na sessão X.” |
| Procedural | Como executar ou evitar uma ação | “Validar migrations antes do deploy.” |
| Preferência | Escolha persistente de usuário, cliente ou equipe | “O cliente prefere contato por e-mail.” |
| Personalidade | Característica estável observada | “O usuário prefere respostas objetivas.” |
| Thread | Linha de trabalho em andamento | “A migração fiscal aguarda homologação.” |
| Task | Objetivo ou atividade operacional | “Reprocessar os eventos do lote 42.” |
| Emoção | Sinal afetivo associado a uma experiência | “O cliente demonstrou forte insatisfação.” |

Tipo cognitivo, categoria, importância e proveniência são dimensões complementares.

### Memória associativa

A memória associativa emerge das relações entre memórias e entidades. A projeção FalkorDB permite percorrer essas relações, executar GraphRAG e propagar ativação com decay por salto.

## Formação da memória

O caminho principal executa, conforme configuração:

1. idempotência pela referência da origem;
2. remoção ou mascaramento de secrets e PII antes do armazenamento;
3. compressão LLM em facts, concepts e narrative;
4. geração de SimHash e deduplicação escopada;
5. classificação de categoria, importância e tipo;
6. geração de embedding;
7. definição de decay, validade e proveniência;
8. persistência canônica no PostgreSQL;
9. criação de versão e auditoria;
10. extrações assíncronas de triplets e eventos.

Sem LLM ou embedding, o CRUD canônico continua disponível, porém com menos enriquecimento.

## Recuperação e recall

### Busca exata

Seleciona por identidade, como `source_reference` ou metadata. É apropriada para auditoria e idempotência e não envolve similaridade.

### Busca lexical

Usa full-text search no PostgreSQL sobre conteúdo e resumo. É também o fallback quando embedding ou FalkorDB não estão disponíveis.

### Busca vetorial

Compara embeddings no FalkorDB. Depende de provider de embeddings e de uma projeção atualizada do grafo.

### Busca associativa

Parte de memórias-semente e percorre relacionamentos, recuperando conhecimento conectado mesmo quando ele não repete os termos da consulta.

### Busca temporal

Interpreta janelas como “ontem”, “esta semana” e “últimos 7 dias” em português e inglês e consulta `recorded_at`.

### Ranking híbrido

Combina similaridade, overlap lexical, proximidade de grafo quando fornecida, recência, tags, importância, decay, peso emocional e feedback. Memórias expiradas, removidas ou supersedadas são excluídas do recall normal.

## Interceptação de prompts

O interceptador pode enriquecer um prompt antes de ele chegar ao agente:

1. executa um quick-check;
2. opcionalmente pede a um LLM uma decisão de relevância;
3. busca por vetor, grafo, texto e tempo;
4. remove memórias inativas;
5. prioriza resultados e respeita um budget;
6. enquadra memórias como dados não confiáveis;
7. sanitiza padrões conhecidos de prompt injection;
8. mascara PII;
9. devolve o prompt enriquecido e suas referências.

Hindsight notes podem recuperar incidentes e resoluções quando o prompt contém sinais de erro.

## Ciclo de vida

- **Feedback:** helpful/not-helpful, acessos e injeções medem utilidade.
- **Reconciliação:** extrai fatos e decide ADD, UPDATE, DELETE ou NONE.
- **Consolidação:** combina memórias relacionadas preservando informação única.
- **Reflexão:** transforma sessões e observações em conhecimento reutilizável.
- **Conflito:** compara memórias e permite supersede ou dismiss com decisão humana.
- **Retenção:** expira por validade, política ou baixo valor, com dry-run e limites.
- **Erasure:** remove dados canônicos e projeções derivadas, emitindo recibos.

Esquecer deve encerrar validade ou remover de forma governada, sem destruir silenciosamente a história necessária para auditoria.

## Confiança e proveniência

O trust score combina confiança da origem, validação humana, feedback, idade e supersessão. O resultado inclui score, label e razões legíveis. Isso permite distinguir memória recuperável de memória confiável: relevância e veracidade não são a mesma coisa.

## Persistência

### PostgreSQL — fonte canônica

Responsável por memórias, histórico bi-temporal, tags, relacionamentos curados, versões, transactional outbox, tenants, usuários, auditoria, sessões, notas, decisões, políticas, eventos, retenção e recibos de erasure.

### FalkorDB — projeção derivada

Responsável por consultas vetoriais e de grafo. Nós e arestas devem ser reconstruíveis a partir do PostgreSQL.

### Visualizações de grafo

- O Grafo Global parte das memórias e relações canônicas do PostgreSQL, acrescenta relações derivadas projetadas no FalkorDB e calcula comunidades sobre o conjunto filtrado que está sendo exibido.
- O Ego-grafo explora até quatro saltos ao redor de uma memória, combinando relações curadas/canônicas e caminhos derivados. Ausência de vizinhos é distinguida de indisponibilidade da projeção.
- O Grafo Temporal lê `memory_history`: cada ponto é uma versão histórica, linhas tracejadas representam mudanças da mesma memória e setas de supersessão conectam memórias distintas.
- PostgreSQL permanece a fonte de verdade; a falta do FalkorDB degrada relações derivadas, mas não elimina relações canônicas nem o histórico temporal.

### Redis — cache e coordenação

Mantém cache, rate limits e filas assíncronas. Sua perda não deve eliminar conhecimento canônico.

### EmbeddedStore — modo reduzido

O backend JSON embedded oferece CRUD e busca textual simples para desenvolvimento e demonstração. Não é funcionalmente equivalente ao pipeline de produção.

## Interfaces

- REST API;
- MCP JSON-RPC 2.0, SSE e batch;
- Semantic API: remember, recall, improve e forget;
- painel administrativo React;
- CLI e TUI;
- webhooks e conectores;
- export de proveniência e eval capture.

## Segurança e multi-tenancy

- JWT e API keys de serviço presas ao tenant;
- contexto de tenant nas queries canônicas;
- RBAC para administração;
- rate limiting e CORS;
- trust levels para chamadas locais, subagentes e remotas;
- privacy stripping antes de persistir;
- PII masking antes da injeção;
- framing e sanitização no interceptador;
- erasure no armazenamento canônico e nas projeções.

## Dependências e degradação

| Componente | Comportamento sem ele |
|---|---|
| PostgreSQL | O backend principal não opera; é o system of record |
| FalkorDB | CRUD e texto continuam; vetor, GraphRAG e ativação ficam indisponíveis |
| Redis | Cache e fila durável degradam ou ficam indisponíveis |
| LLM provider | CRUD e retrieval determinístico continuam; extração e reflexão degradam |
| Embedding provider | Busca vetorial não é formada; busca lexical continua |

## Estado das capacidades

### Disponíveis no caminho principal

- persistência multi-tenant e CRUD;
- tags, metadata, proveniência e validade;
- busca exata e lexical;
- score híbrido e filtros de memória ativa;
- feedback, revisão e trust score;
- REST, MCP, CLI, TUI e painel administrativo;
- retenção, erasure e reconstrução de derivados.

### Condicionais

- classificação e compressão LLM;
- embeddings e busca vetorial;
- GraphRAG, comunidades e spreading activation;
- reflexão, reconciliação e conflito semântico;
- fila durável e cache Redis.

### Limites operacionais e pontos de evolução

1. Create/update/delete tentam atualizar o FalkorDB imediatamente e também gravam uma transactional outbox. O worker faz retry com backoff e sempre projeta o estado canônico mais recente, impedindo que eventos antigos ressuscitem nós removidos.
2. A consulta `as-of` reconstrói snapshots por tempo de sistema (`system_from/system_to`) e aplica simultaneamente a validade de negócio (`valid_from/valid_to`). A migration inicializa o histórico das linhas existentes a partir do primeiro estado ainda observável.
3. Memória, snapshot histórico, versão, auditoria e evento de projeção são persistidos na mesma transação PostgreSQL. Projeções derivadas continuam deliberadamente fora dessa transação.
4. Os pipelines auditados que enviam conteúdo externo ou armazenado ao LLM usam sanitizer e framing; novos pipelines devem preservar obrigatoriamente esse trust boundary.
5. Similaridade Jaccard não é tratada como prova de contradição: a rotina automática fica desabilitada por padrão e, quando habilitada, só supersede duplicatas normalizadas exatas.
6. A busca full-text usa configuração `simple`, adequada ao conteúdo multilíngue sem stemming específico; qualidade linguística avançada em pt-BR ainda depende de avaliação e estratégia dedicada.
7. O ranking híbrido usa distância real de um hop nos relacionamentos diretos; relevância multi-hop precisa manter a profundidade explícita para não inflar proximidade.
8. O modo embedded aplica isolamento tenant no CRUD, mas continua sendo uma implementação reduzida e não substitui transparentemente o PostgreSQL/FalkorDB de produção.

## Como comprovar que o produto funciona

A presença de endpoints não é suficiente. A homologação deve medir:

- recall, precision, MRR e NDCG em datasets pt-BR e inglês;
- ausência de vazamento entre tenants;
- exclusão de memórias expiradas e supersedadas;
- recuperação imediata após create, update e delete;
- fidelidade das consultas temporais;
- resiliência sem LLM, Redis ou FalkorDB;
- taxa de contexto útil e de contexto indevido;
- resistência a prompt injection armazenada;
- rebuild e recuperação de desastre;
- consistência entre memória, versão e auditoria.

O produto cumpre sua proposta quando um agente recupera conhecimento correto entre sessões, com escopo, proveniência e validade verificáveis, e quando informação errada pode ser corrigida ou retirada do recall sem destruir a história necessária.
