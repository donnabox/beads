-- Test-only projection of BDP_GRAPH_CLI_AND_STORAGE_SPEC.md B4 row shapes.
-- Not a migration or a minted graph Scope: no authority, ledger or allocations.
-- These four tables support row decoder/query qualification only.
CREATE TABLE graph_scope (
 id TINYINT NOT NULL PRIMARY KEY,
 scope_url VARCHAR(2048) COLLATE utf8mb4_bin NOT NULL,
 authority_id CHAR(32) NOT NULL,
 epoch BIGINT UNSIGNED NOT NULL,
 minted_at DATETIME(6) NOT NULL,
 CHECK (id = 1)
);
CREATE TABLE graph_type_descriptors (
 url VARCHAR(2048) COLLATE utf8mb4_bin NOT NULL PRIMARY KEY,
 descriptor LONGBLOB NOT NULL,
 fingerprint CHAR(64) NOT NULL,
 installed_seq BIGINT UNSIGNED NOT NULL,
 installed_at DATETIME(6) NOT NULL,
 last_authority_id CHAR(32) NOT NULL,
 last_epoch BIGINT UNSIGNED NOT NULL,
 UNIQUE (fingerprint)
);
CREATE TABLE graph_beads (
 path VARCHAR(1024) COLLATE utf8mb4_bin NOT NULL PRIMARY KEY,
 type_url VARCHAR(2048) COLLATE utf8mb4_bin NOT NULL,
 revision CHAR(32) NOT NULL,
 attribution_principal VARCHAR(512) NULL,
 attribution_status ENUM('claimed','unknown') NULL,
 properties LONGBLOB NOT NULL,
 last_authority_id CHAR(32) NOT NULL,
 last_epoch BIGINT UNSIGNED NOT NULL,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 INDEX (type_url, path),
 FOREIGN KEY (type_url) REFERENCES graph_type_descriptors(url),
 CHECK ((attribution_principal IS NULL AND attribution_status IS NULL) OR (attribution_principal IS NOT NULL AND attribution_status IS NOT NULL))
);
CREATE TABLE graph_links (
 path VARCHAR(1024) COLLATE utf8mb4_bin NOT NULL PRIMARY KEY,
 type_url VARCHAR(2048) COLLATE utf8mb4_bin NOT NULL,
 revision CHAR(32) NOT NULL,
 attribution_principal VARCHAR(512) NULL,
 attribution_status ENUM('claimed','unknown') NULL,
 properties LONGBLOB NOT NULL,
 source_kind ENUM('in','ext') NOT NULL,
 source_path VARCHAR(1024) COLLATE utf8mb4_bin NULL,
 source_url VARCHAR(2048) COLLATE utf8mb4_bin NULL,
 source_pin VARCHAR(512) COLLATE utf8mb4_bin NULL,
 target_kind ENUM('in','ext') NOT NULL,
 target_path VARCHAR(1024) COLLATE utf8mb4_bin NULL,
 target_url VARCHAR(2048) COLLATE utf8mb4_bin NULL,
 target_pin VARCHAR(512) COLLATE utf8mb4_bin NULL,
 last_authority_id CHAR(32) NOT NULL,
 last_epoch BIGINT UNSIGNED NOT NULL,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 INDEX (source_path, type_url, path),
 INDEX (target_path, type_url, path),
 FOREIGN KEY (source_path) REFERENCES graph_beads(path),
 CHECK ((source_kind = 'in' AND source_path IS NOT NULL AND source_url IS NULL) OR (source_kind = 'ext' AND source_path IS NULL AND source_url IS NOT NULL)),
 CHECK ((target_kind = 'in' AND target_path IS NOT NULL AND target_url IS NULL) OR (target_kind = 'ext' AND target_path IS NULL AND target_url IS NOT NULL)),
 CHECK (source_kind = 'in' OR target_kind = 'in'),
 CHECK ((attribution_principal IS NULL AND attribution_status IS NULL) OR (attribution_principal IS NOT NULL AND attribution_status IS NOT NULL))
);
