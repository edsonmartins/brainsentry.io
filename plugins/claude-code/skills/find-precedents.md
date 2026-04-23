---
name: find-precedents
description: Busca decisões passadas semelhantes a um cenário atual antes de registrar uma nova. Use ao ponderar escolhas recorrentes (reviews, refactors, escolhas de dependência).
---

# Find Precedents

Dois modos:

1. Por descrição livre — `POST /v1/decisions/precedents`:

```json
{
  "category": "code_review",
  "scenario": "escolhendo entre zustand e redux",
  "limit": 5
}
```

2. Por decisão existente — `GET /v1/decisions/{id}/precedents?limit=5`.

Apresente cada precedente com similarity, outcome e link ao causal-chain se relevante.
