-- Test-only B4 column/key projection extending fixture.sql to eight tables.
-- It is not a migration, lawful ledger, witness, or full E2/E4 schema evidence.
-- Per-event-kind ledger checks and production fencing triggers are not installed.
CREATE TABLE graph_scope_history (
 scope_url VARCHAR(2048) COLLATE utf8mb4_bin NOT NULL PRIMARY KEY,
 refused_seq BIGINT UNSIGNED NOT NULL,
 refused_at DATETIME(6) NOT NULL,
 reason VARCHAR(64) NOT NULL
);
CREATE TABLE graph_ledger_seq (
 id TINYINT NOT NULL PRIMARY KEY,
 next_seq BIGINT UNSIGNED NOT NULL,
 alloc_nonce CHAR(32) NOT NULL
);
CREATE TABLE graph_ledger_events (
 seq BIGINT UNSIGNED NOT NULL PRIMARY KEY,
 op_id CHAR(32) NOT NULL,
 kind ENUM('mint','install','update','promote','rotate','allocate','tombstone','refuse_url') NOT NULL,
 path VARCHAR(1024) COLLATE utf8mb4_bin NULL,
 scope_url VARCHAR(2048) COLLATE utf8mb4_bin NULL,
 resource_kind ENUM('bead','link') NULL,
 revision CHAR(32) NULL,
 state ENUM('pruned','erased') NULL,
 fingerprint CHAR(64) NULL,
 authority_id CHAR(32) NOT NULL,
 epoch BIGINT UNSIGNED NOT NULL,
 at DATETIME(6) NOT NULL,
 prev_hash CHAR(64) NOT NULL,
 hash CHAR(64) NOT NULL,
 UNIQUE (hash),
 INDEX (path, seq),
 INDEX (op_id)
);
CREATE TABLE graph_authority_lease (
 id TINYINT NOT NULL PRIMARY KEY,
 scope_url VARCHAR(2048) COLLATE utf8mb4_bin NOT NULL,
 authority_id CHAR(32) NOT NULL,
 holder_installation_key CHAR(64) NOT NULL,
 renewer CHAR(32) NOT NULL,
 epoch BIGINT UNSIGNED NOT NULL,
 granted_at DATETIME(6) NOT NULL,
 expires_at DATETIME(6) NOT NULL,
 heartbeat_at DATETIME(6) NOT NULL,
 fence CHAR(32) NOT NULL,
 CHECK (id = 1)
);
INSERT INTO dolt_ignore (pattern, ignored) VALUES ('graph_authority_lease', true);
-- Deliberately minimal unrelated issue-plane control, not the issue schema.
CREATE TABLE issues (id VARCHAR(64) PRIMARY KEY, title TEXT NOT NULL);
INSERT INTO graph_ledger_seq VALUES (0, 1, REPEAT('a', 32));
INSERT INTO graph_authority_lease VALUES
 (1, 'https://graph.example/fixture/', REPEAT('a', 32), REPEAT('b', 64),
  REPEAT('c', 32), 1, '2026-09-17 00:00:00', '2026-09-17 01:00:00',
  '2026-09-17 00:00:00', REPEAT('d', 32));
