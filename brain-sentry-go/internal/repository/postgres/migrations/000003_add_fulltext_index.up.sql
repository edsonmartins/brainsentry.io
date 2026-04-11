-- Add GIN index for full-text search on memories
CREATE INDEX IF NOT EXISTS idx_memories_fulltext ON memories
  USING GIN (to_tsvector('english', coalesce(content,'') || ' ' || coalesce(summary,'')));
