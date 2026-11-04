CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    code VARCHAR(8) NOT NULL UNIQUE,
    original_url TEXT NOT NULL
);

CREATE INDEX idx_links_code ON links(code);

CREATE INDEX idx_links_original_url ON links(original_url);