-- +migrate Up notransaction
CREATE INDEX CONCURRENTLY ON test_reports(app_slug, build_slug, uploaded);

-- +migrate Down
DROP INDEX test_reports_app_slug_build_slug_uploaded_idx;
