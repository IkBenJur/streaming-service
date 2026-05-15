-- +goose Up
ALTER TABLE videos ALTER COLUMN created_at SET DEFAULT NOW();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER videos_set_updated_at
    BEFORE UPDATE ON videos
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS videos_set_updated_at ON videos;
DROP FUNCTION IF EXISTS set_updated_at();
ALTER TABLE videos ALTER COLUMN created_at DROP DEFAULT;
