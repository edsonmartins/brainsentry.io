---
name: as-of
description: Consulta a BrainSentry como ela era em um instante passado (time-travel). Use em compliance, post-mortem ou quando o usuário pergunta "o que sabíamos em <data>?".
---

# As-Of

Chame `GET /v1/memories/as-of?at=<RFC3339>&limit=100`.

O endpoint retorna apenas memórias cujo `valid_from <= at < valid_until` **e** que já haviam sido gravadas (`recorded_at <= at`). Isso preserva a linha do tempo bi-temporal e é o alicerce de auditoria para decisões retroativas.

Formate a saída em blocos por categoria com os campos: id, summary, createdAt, validFrom/validUntil.
