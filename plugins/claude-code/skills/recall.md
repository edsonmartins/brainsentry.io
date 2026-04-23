---
name: recall
description: Recupera memórias relevantes da BrainSentry para o contexto da conversa. Use quando a tarefa depende de conhecimento anterior do usuário ou do projeto.
---

# Recall

Quando ativada, consulte o endpoint `/v1/recall` via MCP (ou `/v1/memories/search` como fallback) passando a pergunta atual. Monte o corpo assim:

```json
{
  "query": "<pergunta curta>",
  "topK": 5
}
```

Apresente cada memória com seu id, resumo, importância e score. Cite IDs ao usar uma memória em sua resposta.
