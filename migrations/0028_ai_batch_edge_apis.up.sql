-- 0028: workspace resources — AI hosting batch/edge + AI APIs vision/language/speech.
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── AI Batch Inference Jobs ────────────────────────────────────────────────────
-- Covers ai/hosting/batch
CREATE TABLE IF NOT EXISTS workspace_ai_batch_jobs (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    model        TEXT NOT NULL DEFAULT '',
    input_path   TEXT NOT NULL DEFAULT '',
    output_path  TEXT NOT NULL DEFAULT '',
    records      TEXT NOT NULL DEFAULT '—',
    processed    TEXT NOT NULL DEFAULT '0',
    progress     TEXT NOT NULL DEFAULT '0%',
    runtime      TEXT NOT NULL DEFAULT '—',
    cost         TEXT NOT NULL DEFAULT '$0'
);

CREATE INDEX IF NOT EXISTS workspace_ai_batch_jobs_resource_idx
    ON workspace_ai_batch_jobs (resource_id);

-- ── AI Edge Deployments ────────────────────────────────────────────────────────
-- Covers ai/hosting/edge
-- runtime: 'ONNX Runtime' | 'TensorRT' | 'TFLite' | 'OpenVINO'
CREATE TABLE IF NOT EXISTS workspace_ai_edge_deployments (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    model        TEXT NOT NULL DEFAULT '',
    runtime      TEXT NOT NULL DEFAULT 'ONNX Runtime',
    target       TEXT NOT NULL DEFAULT '',
    devices      INTEGER NOT NULL DEFAULT 0,
    version      TEXT NOT NULL DEFAULT 'v0.1.0',
    latency      TEXT NOT NULL DEFAULT '—'
);

CREATE INDEX IF NOT EXISTS workspace_ai_edge_deployments_resource_idx
    ON workspace_ai_edge_deployments (resource_id);

-- ── AI Vision APIs ─────────────────────────────────────────────────────────────
-- Covers ai/apis/vision
-- capability: 'object-detection' | 'classification' | 'ocr' | 'face-detection' | 'segmentation'
CREATE TABLE IF NOT EXISTS workspace_ai_vision_apis (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    model        TEXT NOT NULL DEFAULT '',
    capability   TEXT NOT NULL DEFAULT '',
    version      TEXT NOT NULL DEFAULT 'v1.0.0',
    requests     TEXT NOT NULL DEFAULT '0/mo',
    latency      TEXT NOT NULL DEFAULT '—',
    accuracy     TEXT NOT NULL DEFAULT '—',
    endpoint     TEXT NOT NULL DEFAULT '',
    api_key      TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS workspace_ai_vision_apis_resource_idx
    ON workspace_ai_vision_apis (resource_id);

-- ── AI Language APIs ───────────────────────────────────────────────────────────
-- Covers ai/apis/language
-- capability: 'sentiment' | 'translation' | 'summarization' | 'ner' | 'question-answering'
CREATE TABLE IF NOT EXISTS workspace_ai_language_apis (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    model        TEXT NOT NULL DEFAULT '',
    capability   TEXT NOT NULL DEFAULT '',
    version      TEXT NOT NULL DEFAULT 'v1.0.0',
    requests     TEXT NOT NULL DEFAULT '0/mo',
    latency      TEXT NOT NULL DEFAULT '—',
    languages    TEXT NOT NULL DEFAULT 'en',
    endpoint     TEXT NOT NULL DEFAULT '',
    api_key      TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS workspace_ai_language_apis_resource_idx
    ON workspace_ai_language_apis (resource_id);

-- ── AI Speech APIs ─────────────────────────────────────────────────────────────
-- Covers ai/apis/speech
-- capability: 'stt' | 'tts' | 'speaker-id' | 'voice-clone'
CREATE TABLE IF NOT EXISTS workspace_ai_speech_apis (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    model        TEXT NOT NULL DEFAULT '',
    capability   TEXT NOT NULL DEFAULT '',
    version      TEXT NOT NULL DEFAULT 'v1.0.0',
    requests     TEXT NOT NULL DEFAULT '0/mo',
    latency      TEXT NOT NULL DEFAULT '—',
    languages    TEXT NOT NULL DEFAULT 'en',
    sample_rate  TEXT NOT NULL DEFAULT '16kHz',
    endpoint     TEXT NOT NULL DEFAULT '',
    api_key      TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS workspace_ai_speech_apis_resource_idx
    ON workspace_ai_speech_apis (resource_id);
