CREATE SCHEMA publication;

-- append-only
CREATE TABLE publication.receipt (
  row_id SERIAL PRIMARY KEY,
  transaction_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  envelope_key BYTEA UNIQUE NOT NULL,
  entry_key BYTEA NOT NULL,
  author_public_key BYTEA NOT NULL,
  author_entity_id VARCHAR,
  reader_public_key BYTEA NOT NULL,
  reader_entity_id VARCHAR,
  received_time BIGINT,
  received_date DATE
);

-- support filters on all of these fields
CREATE INDEX receipt_entry_key ON publication.receipt (entry_key);
CREATE INDEX receipt_author_public_key ON publication.receipt (author_public_key);
CREATE INDEX receipt_author_entity_id ON publication.receipt (author_entity_id);
CREATE INDEX receipt_reader_public_key ON publication.receipt (reader_public_key);
CREATE INDEX receipt_reader_entity_id ON publication.receipt (reader_entity_id);
CREATE INDEX receipt_received_date ON publication.receipt (received_date);
