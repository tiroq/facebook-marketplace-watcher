-- Extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- publishers
CREATE TABLE IF NOT EXISTS publishers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source text NOT NULL,
    source_publisher_id text,
    display_name text,
    profile_url text,
    profile_url_hash text,
    location_hint text,
    raw_json jsonb NOT NULL DEFAULT '{}',
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS publishers_source_source_publisher_id_uidx
    ON publishers (source, source_publisher_id)
    WHERE source_publisher_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS publishers_source_idx ON publishers (source);
CREATE INDEX IF NOT EXISTS publishers_source_publisher_id_idx ON publishers (source_publisher_id);
CREATE INDEX IF NOT EXISTS publishers_profile_url_hash_idx ON publishers (profile_url_hash);

-- listings
CREATE TABLE IF NOT EXISTS listings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source text NOT NULL,
    source_listing_id text NOT NULL,
    publisher_id uuid REFERENCES publishers(id),
    canonical_url text,
    canonical_url_hash text,
    title_current text,
    price_current numeric(14,2),
    currency text,
    location_current text,
    status text NOT NULL DEFAULT 'active',
    detail_status text NOT NULL DEFAULT 'not_requested',
    enrichment_status text NOT NULL DEFAULT 'not_requested',
    duplicate_status text NOT NULL DEFAULT 'unique',
    product_cluster_id uuid NULL,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    version bigint NOT NULL DEFAULT 1,
    raw_json jsonb NOT NULL DEFAULT '{}',
    UNIQUE (source, source_listing_id)
);
CREATE INDEX IF NOT EXISTS listings_source_idx ON listings (source);
CREATE INDEX IF NOT EXISTS listings_source_listing_id_idx ON listings (source_listing_id);
CREATE INDEX IF NOT EXISTS listings_publisher_id_idx ON listings (publisher_id);
CREATE INDEX IF NOT EXISTS listings_last_seen_at_idx ON listings (last_seen_at);
CREATE INDEX IF NOT EXISTS listings_status_idx ON listings (status);
CREATE INDEX IF NOT EXISTS listings_title_current_trgm_idx ON listings USING gin (title_current gin_trgm_ops);

-- search_queries
CREATE TABLE IF NOT EXISTS search_queries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    query text NOT NULL,
    source text NOT NULL DEFAULT 'facebook_marketplace',
    location_hint text,
    enabled boolean NOT NULL DEFAULT true,
    max_cards_per_run int NOT NULL DEFAULT 50,
    max_scrolls_per_run int NOT NULL DEFAULT 5,
    active_window_start time,
    active_window_end time,
    timezone text NOT NULL DEFAULT 'Asia/Bangkok',
    interval_minutes int NOT NULL DEFAULT 40,
    jitter_minutes int NOT NULL DEFAULT 10,
    skip_probability numeric(5,4) NOT NULL DEFAULT 0.5,
    max_runs_per_day int NOT NULL DEFAULT 3,
    raw_policy jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- search_sessions
CREATE TABLE IF NOT EXISTS search_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    search_query_id uuid REFERENCES search_queries(id),
    status text NOT NULL DEFAULT 'created',
    started_at timestamptz,
    finished_at timestamptz,
    error_code text,
    error_message text,
    raw_json jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- listing_observations
CREATE TABLE IF NOT EXISTS listing_observations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id uuid NOT NULL REFERENCES listings(id),
    search_session_id uuid REFERENCES search_sessions(id),
    search_query_id uuid REFERENCES search_queries(id),
    observed_at timestamptz NOT NULL DEFAULT now(),
    position int,
    title_observed text,
    price_observed numeric(14,2),
    price_text text,
    currency text,
    location_observed text,
    url_observed text,
    publisher_hint text,
    raw_text text,
    raw_json jsonb NOT NULL DEFAULT '{}',
    screenshot_path text,
    html_hash text,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS listing_observations_listing_id_idx ON listing_observations (listing_id);
CREATE INDEX IF NOT EXISTS listing_observations_search_session_id_idx ON listing_observations (search_session_id);
CREATE INDEX IF NOT EXISTS listing_observations_search_query_id_idx ON listing_observations (search_query_id);
CREATE INDEX IF NOT EXISTS listing_observations_observed_at_idx ON listing_observations (observed_at);
CREATE INDEX IF NOT EXISTS listing_observations_title_observed_trgm_idx ON listing_observations USING gin (title_observed gin_trgm_ops);

-- listing_detail_snapshots
CREATE TABLE IF NOT EXISTS listing_detail_snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id uuid NOT NULL REFERENCES listings(id),
    captured_at timestamptz NOT NULL DEFAULT now(),
    url text,
    description_raw text,
    description_text text,
    seller_text text,
    attributes_raw_json jsonb NOT NULL DEFAULT '{}',
    image_urls_json jsonb NOT NULL DEFAULT '[]',
    html_snapshot_path text,
    screenshot_path text,
    parse_version text NOT NULL DEFAULT 'v1',
    raw_json jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now()
);

-- listing_images
CREATE TABLE IF NOT EXISTS listing_images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id uuid NOT NULL REFERENCES listings(id),
    source_image_url text,
    source_image_url_hash text,
    local_path text,
    sha256 text,
    phash text,
    width int,
    height int,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS listing_images_listing_id_idx ON listing_images (listing_id);
CREATE INDEX IF NOT EXISTS listing_images_source_image_url_hash_idx ON listing_images (source_image_url_hash);
CREATE INDEX IF NOT EXISTS listing_images_sha256_idx ON listing_images (sha256);
CREATE INDEX IF NOT EXISTS listing_images_phash_idx ON listing_images (phash);

-- product_clusters
CREATE TABLE IF NOT EXISTS product_clusters (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    category text,
    subcategory text,
    manufacturer text,
    model text,
    representative_listing_id uuid REFERENCES listings(id),
    cluster_confidence numeric(5,4),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- duplicate_candidates
CREATE TABLE IF NOT EXISTS duplicate_candidates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id uuid NOT NULL REFERENCES listings(id),
    candidate_listing_id uuid NOT NULL REFERENCES listings(id),
    candidate_type text NOT NULL,
    score numeric(5,4) NOT NULL,
    reason_json jsonb NOT NULL DEFAULT '{}',
    status text NOT NULL DEFAULT 'pending',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (listing_id, candidate_listing_id, candidate_type)
);

-- listing_metadata
CREATE TABLE IF NOT EXISTS listing_metadata (
    listing_id uuid PRIMARY KEY REFERENCES listings(id),
    language text,
    normalized_title_en text,
    normalized_description_en text,
    category text,
    subcategory text,
    manufacturer text,
    model text,
    specs_json jsonb NOT NULL DEFAULT '{}',
    confidence_json jsonb NOT NULL DEFAULT '{}',
    extraction_version text,
    llm_model text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- price_observations
CREATE TABLE IF NOT EXISTS price_observations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id uuid NOT NULL REFERENCES listings(id),
    observed_at timestamptz NOT NULL DEFAULT now(),
    price_amount numeric(14,2),
    currency text,
    source text NOT NULL,
    search_query_id uuid REFERENCES search_queries(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- market_price_baselines
CREATE TABLE IF NOT EXISTS market_price_baselines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    category text,
    subcategory text,
    manufacturer text,
    model text,
    spec_fingerprint text,
    location_scope text,
    currency text,
    sample_size int NOT NULL DEFAULT 0,
    price_min numeric(14,2),
    price_p25 numeric(14,2),
    price_median numeric(14,2),
    price_p75 numeric(14,2),
    price_p90 numeric(14,2),
    price_max numeric(14,2),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- listing_scores
CREATE TABLE IF NOT EXISTS listing_scores (
    listing_id uuid PRIMARY KEY REFERENCES listings(id),
    deal_score numeric(5,4),
    price_score numeric(5,4),
    quality_score numeric(5,4),
    risk_score numeric(5,4),
    freshness_score numeric(5,4),
    fit_score numeric(5,4),
    publisher_score numeric(5,4),
    fair_price_estimate numeric(14,2),
    discount_ratio numeric(8,4),
    decision text,
    reason_json jsonb NOT NULL DEFAULT '[]',
    score_version text NOT NULL DEFAULT 'v1',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- llm_requests
CREATE TABLE IF NOT EXISTS llm_requests (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    task_type text NOT NULL,
    input_hash text NOT NULL,
    provider text,
    model text,
    status text NOT NULL DEFAULT 'created',
    request_json jsonb NOT NULL DEFAULT '{}',
    response_json jsonb NOT NULL DEFAULT '{}',
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- task_audit_log
CREATE TABLE IF NOT EXISTS task_audit_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    service text NOT NULL,
    event_id text,
    subject text,
    status text NOT NULL,
    message text,
    raw_json jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now()
);

-- updated_at trigger function
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables with updated_at
CREATE TRIGGER set_updated_at_publishers BEFORE UPDATE ON publishers FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at_listings BEFORE UPDATE ON listings FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at_search_queries BEFORE UPDATE ON search_queries FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at_search_sessions BEFORE UPDATE ON search_sessions FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at_product_clusters BEFORE UPDATE ON product_clusters FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at_duplicate_candidates BEFORE UPDATE ON duplicate_candidates FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at_listing_metadata BEFORE UPDATE ON listing_metadata FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at_listing_scores BEFORE UPDATE ON listing_scores FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at_llm_requests BEFORE UPDATE ON llm_requests FOR EACH ROW EXECUTE FUNCTION set_updated_at();
