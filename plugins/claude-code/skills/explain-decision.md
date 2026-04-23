---
name: explain-decision
description: Explica uma decisão passada via reasoning abdutivo. Use quando o usuário perguntar "por que decidimos X?" ou quiser auditar uma escolha antiga.
---

# Explain Decision

Passo 1 — confirme o id da decisão (pergunte se não tiver).

Passo 2 — chame:

```
POST /v1/reasoning/abduce
{
  "decisionId": "<uuid>",
  "question": "<pergunta opcional>",
  "maxHypotheses": 5
}
```

Passo 3 — apresente as hipóteses ordenadas por confiança e liste IDs das memórias usadas como evidência. Se houver `parentDecisionId`, ofereça recuperar a cadeia causal (`GET /v1/decisions/{id}/causal-chain`).
