-- Use PostgreSQL's language-neutral configuration for the bilingual pt-BR/en
-- product. The previous english dictionary dropped/stemmed Portuguese terms
-- incorrectly and made lexical recall depend on the language of the memory.
DROP INDEX IF EXISTS idx_memories_fulltext;

CREATE INDEX idx_memories_fulltext ON memories
  USING GIN (to_tsvector('simple', coalesce(content,'') || ' ' || coalesce(summary,'')));
