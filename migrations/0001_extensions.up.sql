-- 0001_extensions: Postgres extensions required across the schema.
-- pgcrypto: gen_random_uuid(), digest()
-- citext:   case-insensitive text for emails, slugs
-- pg_trgm:  trigram indexes for fuzzy search on names, SQL console autocomplete

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
