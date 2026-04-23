---
name: remember
description: Salva um fato, decisão ou preferência importante na BrainSentry para reuso futuro. Use após aprender algo durável sobre o usuário ou projeto.
---

# Remember

Chame o endpoint `/v1/remember` via MCP com:

```json
{
  "content": "<texto>",
  "category": "<opcional>",
  "importance": "<low|medium|high|critical>",
  "tags": ["..."]
}
```

Reporte o `memoryId` retornado para o usuário. Não salve segredos, tokens ou PII — o pipeline tira, mas você deve evitar a entrada.
