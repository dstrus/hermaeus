-- The tables Hermaeus reads and writes in a SkyrimNet database, for the tests:
-- each table's columns with their types, keys, and defaults, as of SkyrimNet
-- Beta 28. Only what Hermaeus relies on; indexes that only speed up queries
-- are left out.

CREATE TABLE uuid_mappings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid BIGINT NOT NULL,
  form_id INTEGER NOT NULL,
  actor_name TEXT,
  bio_template_name TEXT,
  created_at REAL DEFAULT (julianday('now')),
  updated_at REAL DEFAULT (julianday('now')),
  mod_file TEXT DEFAULT NULL,
  local_form_id INTEGER DEFAULT NULL,
  UNIQUE (uuid)
);

CREATE TABLE identity_aliases (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  old_uuid BIGINT NOT NULL,
  new_uuid BIGINT NOT NULL,
  mod_file TEXT DEFAULT NULL,
  local_form_id INTEGER DEFAULT NULL,
  old_name TEXT DEFAULT NULL,
  new_name TEXT DEFAULT NULL,
  game_time REAL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE memories (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  actor_uuid BIGINT NOT NULL,
  content TEXT NOT NULL,
  location TEXT,
  game_time REAL,
  creation_time REAL DEFAULT (julianday('now')),
  related_event_ids TEXT,
  related_actors TEXT,
  emotion TEXT,
  importance_score REAL DEFAULT 0.5,
  tags TEXT,
  memory_type TEXT DEFAULT 'EXPERIENCE',
  embedding_checksum TEXT,
  condition_expr TEXT DEFAULT NULL,
  always_inject INTEGER DEFAULT 0,
  is_active INTEGER DEFAULT 1,
  display_name TEXT DEFAULT NULL,
  pack_id INTEGER DEFAULT NULL REFERENCES knowledge_packs(id) ON DELETE SET NULL,
  knowledge_key TEXT DEFAULT NULL,
  source_key TEXT DEFAULT NULL,
  source_hash TEXT DEFAULT NULL,
  updated_at REAL
);

CREATE TABLE memory_embeddings (
  memory_id INTEGER NOT NULL,
  embedding BLOB NOT NULL,
  dimension INTEGER NOT NULL,
  UNIQUE (memory_id)
);

CREATE TABLE diary_entries (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  actor_uuid BIGINT NOT NULL,
  content TEXT NOT NULL,
  entry_date REAL,
  creation_time REAL DEFAULT (julianday('now')),
  location TEXT,
  emotion TEXT,
  related_event_ids TEXT,
  related_memory_ids TEXT,
  importance_score REAL DEFAULT 0.5,
  tags TEXT,
  updated_at REAL
);

CREATE TABLE events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_type TEXT NOT NULL,
  event_data TEXT NOT NULL,
  location TEXT,
  game_time REAL,
  game_time_str TEXT,
  local_time REAL DEFAULT (julianday('now')),
  actor_UUIDs TEXT,
  originating_actor_UUID BIGINT DEFAULT 0,
  target_actor_UUID BIGINT DEFAULT 0,
  memory_generated_for_actors TEXT DEFAULT '[]',
  playtime INTEGER
);

CREATE TABLE knowledge_packs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT DEFAULT '',
  version TEXT DEFAULT '1.0.0',
  is_enabled INTEGER DEFAULT 1,
  created_at TEXT DEFAULT (datetime('now')),
  updated_at TEXT DEFAULT (datetime('now')),
  source_path TEXT DEFAULT NULL,
  source_file_hash TEXT DEFAULT NULL,
  UNIQUE (source_path)
);

CREATE TABLE omnisight_screenshots (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  subject_type TEXT NOT NULL,
  subject_identifier TEXT NOT NULL,
  subject_name TEXT,
  location_name TEXT,
  location_cell TEXT,
  position_x REAL,
  position_y REAL,
  position_z REAL,
  capture_time_local TEXT NOT NULL,
  capture_time_game TEXT,
  capture_timestamp_local REAL NOT NULL,
  capture_timestamp_game REAL,
  file_path TEXT,
  image_format TEXT DEFAULT 'png',
  image_width INTEGER,
  image_height INTEGER,
  image_size_bytes INTEGER,
  description TEXT,
  description_model TEXT,
  description_prompt TEXT,
  description_timestamp REAL,
  quality_score REAL DEFAULT 0.0,
  line_of_sight_quality REAL,
  distance_to_subject REAL,
  is_indoor BOOLEAN DEFAULT 0,
  weather_type TEXT,
  time_of_day TEXT,
  is_valid BOOLEAN DEFAULT 1,
  needs_refresh BOOLEAN DEFAULT 0,
  generation_error TEXT,
  version INTEGER DEFAULT 1,
  content_hash TEXT,
  suggested_name TEXT,
  description_locked BOOLEAN DEFAULT 0,
  manual_capture BOOLEAN DEFAULT 0,
  UNIQUE (subject_type, subject_identifier)
);

CREATE TABLE bard_songs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  save_id TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL,
  lyrics TEXT NOT NULL,
  style_tags TEXT NOT NULL,
  gender TEXT NOT NULL,
  audio_filename TEXT NOT NULL,
  suno_song_id TEXT,
  duration_seconds REAL,
  game_time_created REAL NOT NULL,
  play_count INTEGER DEFAULT 0,
  is_cached INTEGER DEFAULT 1,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  composer_uuid INTEGER NOT NULL DEFAULT 0,
  composer_name TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT 'generated',
  record_id TEXT NOT NULL DEFAULT '',
  instrumental INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE npc_groups (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  group_name TEXT NOT NULL,
  description TEXT DEFAULT '',
  UNIQUE (group_name)
);

CREATE TABLE npc_group_members (
  group_id INTEGER NOT NULL REFERENCES npc_groups(id) ON DELETE CASCADE,
  actor_uuid BIGINT NOT NULL,
  PRIMARY KEY (group_id, actor_uuid)
);

CREATE TABLE embedding_producer (
  id INTEGER PRIMARY KEY,
  fingerprint TEXT NOT NULL,
  dimension INTEGER NOT NULL
);

-- SkyrimNet's own bookkeeping, which Hermaeus doesn't read.
CREATE TABLE schema_migrations (
  version INTEGER PRIMARY KEY
);

CREATE TABLE dialogue (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  speaker TEXT NOT NULL,
  text TEXT NOT NULL,
  timestamp REAL DEFAULT (julianday('now'))
);

CREATE TABLE traces (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  timestamp REAL,
  depth INTEGER,
  label TEXT,
  duration_us REAL,
  parent_label TEXT,
  session_id TEXT
);

CREATE TABLE counters (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  label TEXT,
  value INTEGER,
  session_id TEXT,
  UNIQUE (label, session_id)
);

CREATE TABLE errors (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  label TEXT,
  value INTEGER,
  session_id TEXT,
  UNIQUE (label, session_id)
);

CREATE UNIQUE INDEX idx_knowledge_packs_source_path ON knowledge_packs(source_path) WHERE source_path IS NOT NULL;
