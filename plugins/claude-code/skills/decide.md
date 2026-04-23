---
name: decide
description: Registra uma decisão auditável na BrainSentry com reasoning, outcome e memórias citadas. Use quando você decidir algo não trivial em nome do usuário.
---

# Decide

Chame `POST /v1/decisions` com:

```json
{
  "category": "<ex.: code_review, refactor, architecture>",
  "scenario": "<descrição curta da situação>",
  "reasoning": "<por que esta escolha>",
  "outcome": "approved|rejected|deferred|pending",
  "confidence": 0.85,
  "memoryIds": ["<ids citados>"],
  "entityIds": ["<opcional>"],
  "parentDecisionId": "<opcional>"
}
```

Se a resposta contiver `policyViolations`, destaque-as para o usuário antes de prosseguir.
