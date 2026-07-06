-- Create an extension for vector similarity search if supported
-- Note: Requires pgvector extension in Postgres
CREATE EXTENSION IF NOT EXISTS vector;

-- memory_vault table stores generic nodes of knowledge
CREATE TABLE IF NOT EXISTS memory_nodes (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL, -- e.g., 'Document', 'Interaction', 'Objection'
    content TEXT NOT NULL,
    embedding vector(384), -- 384-dim embedding for lightweight local models (e.g. all-MiniLM-L6-v2)
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- memory_edges table stores relationships for GraphRAG
CREATE TABLE IF NOT EXISTS memory_edges (
    id SERIAL PRIMARY KEY,
    source_node_id INTEGER NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
    target_node_id INTEGER NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
    relation_type VARCHAR(100) NOT NULL, -- e.g., 'RELATES_TO', 'CONTRADICTS', 'SUPPORTS'
    weight FLOAT DEFAULT 1.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_node_id, target_node_id, relation_type)
);

-- Indexes for fast traversal and vector search
CREATE INDEX ON memory_nodes USING hnsw (embedding vector_cosine_ops);
CREATE INDEX ON memory_edges(source_node_id);
CREATE INDEX ON memory_edges(target_node_id);
