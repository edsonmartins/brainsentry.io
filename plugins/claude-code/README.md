# BrainSentry — Claude Code plugin

Native plugin that exposes BrainSentry cognitive memory + decision intelligence
to Claude Code via MCP.

## Configure

1. Ensure BrainSentry backend is running (defaults to `http://localhost:8082`).
2. Export the variables Claude Code expands when loading `plugin.json`:

```sh
export BRAINSENTRY_URL=http://localhost:8082
export BRAINSENTRY_TOKEN=<JWT or service token>
```

3. Point Claude Code at this folder (e.g., `~/.claude/plugins/brainsentry`) or
   install via your plugin registry.

## Skills

- `recall` — semantic memory lookup.
- `remember` — durable memory write with PII strip pipeline.
- `decide` — record a first-class, auditable Decision.
- `explain-decision` — abductive reasoning over past decisions.
- `find-precedents` — similarity search over decisions by category.
- `as-of` — bi-temporal time-travel query on Memory.

## Endpoints used

All requests hit the BrainSentry HTTP API with `Authorization: Bearer $BRAINSENTRY_TOKEN`.
See `brain-sentry-go/docs/api.md` for the full contract.
