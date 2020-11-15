-- +migrate Up notransaction
CREATE INDEX CONCURRENTLY ON test_report_assets(test_report_id);

-- +migrate Down
DROP INDEX test_report_assets_test_report_id_idx;
