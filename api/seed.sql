-- Wolllama seed data — preloaded into the Docker image
-- so the service starts with a preview model already available.

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    github_id INTEGER DEFAULT 0,
    username TEXT NOT NULL,
    avatar_url TEXT,
    wallet_address TEXT UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS models (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submitter_id INTEGER NOT NULL REFERENCES users(id),
    manifest_obj_id TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    description_md TEXT,
    original_name TEXT,
    tag TEXT,
    total_size INTEGER,
    blob_count INTEGER,
    manifest_json TEXT,
    submitter_address TEXT,
    signature TEXT,
    featured BOOLEAN DEFAULT 0,
    available BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_models_display_name ON models(display_name);
CREATE INDEX IF NOT EXISTS idx_models_submitter ON models(submitter_id);
CREATE INDEX IF NOT EXISTS idx_models_available ON models(available);

INSERT INTO users (id, github_id, username, avatar_url, wallet_address, created_at) VALUES (1, 0, 'anonymous', NULL, NULL, '2026-06-07 11:38:26');
INSERT INTO models (id, submitter_id, manifest_obj_id, display_name, description_md, original_name, tag, total_size, blob_count, manifest_json, submitter_address, signature, featured, available, created_at) VALUES (1, 1, 'PPofW8S3LyWmU-EMXiw9W-nME355Q5g_ZtoOUvXt7goBAQAEAA', 'smollm:135m', 'A compact 135M-parameter language model for efficient inference.', 'smollm:135m', '135m', 91739413, 5, '{"wolllamaVersion":2,"provider":"tatum","name":"smollm:135m","ollamaManifest":{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","digest":"sha256:f590523c855b7d0f2741a9e076d4b663b1f128f2617b7fcd3fe7d7b57ce71d83","size":488},"layers":[{"mediaType":"application/vnd.ollama.image.model","digest":"sha256:eb2c714d40d4b35ba4b8ee98475a06d51d8080a17d2d2a75a23665985c739b94","size":91727296},{"mediaType":"application/vnd.ollama.image.template","digest":"sha256:62fbfd9ed093d6e5ac83190c86eec5369317919f4b149598d2dbb38900e9faef","size":182},{"mediaType":"application/vnd.ollama.image.license","digest":"sha256:cfc7749b96f63bd31c3c42b5c471bf756814053e847c10f3eb003417bc523d30","size":11358},{"mediaType":"application/vnd.ollama.image.params","digest":"sha256:ca7a9654b5469dc2d638456f31a51a03367987c54135c089165752d9eeb08cd7","size":89}]},"blobs":{"sha256:62fbfd9ed093d6e5ac83190c86eec5369317919f4b149598d2dbb38900e9faef":{"single":"-zPlDPhJn7zhD6iD-PNNDjqfJPyh4bUxfOCR7gZO1xoBAQACAA"},"sha256:ca7a9654b5469dc2d638456f31a51a03367987c54135c089165752d9eeb08cd7":{"single":"OUpbCFd9dKawdh42ZY1rxnGyIUu0ZP6rg7G8QUKqx6sBAQACAA"},"sha256:cfc7749b96f63bd31c3c42b5c471bf756814053e847c10f3eb003417bc523d30":{"single":"q9V_hlrkIEy7ArhCyvGrsFFxHtbTrQTYa9WlbzGb85sBAQATAA"},"sha256:eb2c714d40d4b35ba4b8ee98475a06d51d8080a17d2d2a75a23665985c739b94":{"chunks":["Og_eZxoUW9FkMy208TW1egEBQz5OJ79-eTQomV4aanIBAQCWAg","xVcx6DGQA6aVw-iKg8k4XDFc8eGKigC8Pq9KSHrkqgQBAQCWAg"]},"sha256:f590523c855b7d0f2741a9e076d4b663b1f128f2617b7fcd3fe7d7b57ce71d83":{"single":"c69cYycTMrrCHddK1dpH3TEmgUpIW9dyx3ljr8xTD8wBAQACAA"}},"createdAt":"2026-06-05T15:28:19.083618Z"}', '0x3d05dc8034e96a1c1ed45da5edc1fa3acec3c23fb0de9805b9e2faba04c2e54e', 'ANwpYWY+E6y6i4Sqx7ZgGjDkZWVl9M+q6Fg8bs/PDppSqddderk7ABuXCBiodo3lohtbhvOK0aCrzgagA/ssbADBJ9RZ0nB8Lnubi3eLfu/uS745BlZq2VpsIYUAZSCK2A==', 1, 1, '2026-06-07 11:39:13');
