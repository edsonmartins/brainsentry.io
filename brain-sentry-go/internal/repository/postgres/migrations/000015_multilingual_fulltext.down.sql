DROP INDEX IF EXISTS idx_memories_fulltext;

CREATE INDEX idx_memories_fulltext ON memories
  USING GIN (to_tsvector('english', coalesce(content,'') || ' ' || coalesce(summary,'')));
