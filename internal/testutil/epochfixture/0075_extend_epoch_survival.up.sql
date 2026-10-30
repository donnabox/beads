-- Forward epoch-survival extension for the published fork lineage.
-- Fork0070/0071/0072 are frozen; upstream permuted those slots and enlarged
-- its epoch CREATE. Preserve their original hashes and add the survival
-- columns here instead. Source issueops stores explicit root-carry and terminal
-- loss; NULL-extending occupied old addresses would revive previously lost
-- identities. The Go admission guard therefore refuses occupied old/partial
-- tables and unknown cursor/hash lineage BEFORE any migration writes.
--
-- Each DDL operation is independently guarded because Dolt DDL is not
-- transactional across statements. Recognized empty partial shape may pass
-- read-only admission, but ordinary production dirty-table guards still apply:
-- uncommitted DDL debris is safely refused, not auto-committed or reset.
-- Controlled raw-body replay is not proof of automatic production recovery.
-- A completed cursor is admitted only with exact types/index/CHECK enforcement.
-- No row backfill, table rebuild, loss reset or address reallocation occurs.
-- Migration ownership must exclude old writers throughout admission/apply.
-- Runtime driver migrations use these guards. The raw CLI bundle is FRESH
-- INSTALL ONLY and needs its direct-DDL override (pre-2.3 prepared-DDL hazard).
-- This extension has no generic down adapter: removing loss/carry metadata
-- would destroy retained meaning. Do not manually replay it on existing data.

SET @epoch_cf_needs_add = (SELECT IF(COUNT(*) = 0, 1, 0) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'epoch_minted_addresses' AND COLUMN_NAME = 'carried_from');
SET @sql = IF(@epoch_cf_needs_add = 1,
    'ALTER TABLE epoch_minted_addresses ADD COLUMN carried_from VARCHAR(255) NULL',
    'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @epoch_ge_needs_add = (SELECT IF(COUNT(*) = 0, 1, 0) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'epoch_minted_addresses' AND COLUMN_NAME = 'gone_at_epoch');
SET @sql = IF(@epoch_ge_needs_add = 1,
    'ALTER TABLE epoch_minted_addresses ADD COLUMN gone_at_epoch INT NULL',
    'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @epoch_ci_needs_add = (SELECT IF(COUNT(*) = 0, 1, 0) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'epoch_minted_addresses' AND INDEX_NAME = 'idx_epoch_minted_addresses_carried_from');
SET @sql = IF(@epoch_ci_needs_add = 1,
    'CREATE INDEX idx_epoch_minted_addresses_carried_from ON epoch_minted_addresses (carried_from)',
    'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @epoch_ck_needs_add = (SELECT IF(COUNT(*) = 0, 1, 0) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'epoch_minted_addresses' AND CONSTRAINT_NAME = 'ck_epoch_minted_addresses_gone_not_before_mint');
SET @sql = IF(@epoch_ck_needs_add = 1,
    'ALTER TABLE epoch_minted_addresses ADD CONSTRAINT ck_epoch_minted_addresses_gone_not_before_mint CHECK (gone_at_epoch IS NULL OR gone_at_epoch >= minted_epoch)',
    'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
