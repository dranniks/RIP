-- 001_init.sql
-- PostgreSQL schema for Lab #3 REST API (XRF domain)

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    login VARCHAR(64) NOT NULL UNIQUE,
    full_name VARCHAR(128) NOT NULL,
    password_hash VARCHAR(255) NOT NULL DEFAULT '',
    role VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reference_alloy_services (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(120) NOT NULL UNIQUE,
    name VARCHAR(160) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(16) NOT NULL,
    image_file_name VARCHAR(160) NULL,
    video_file_name VARCHAR(160) NULL,
    image_url VARCHAR(255) NULL,
    video_url VARCHAR(255) NULL,
    era VARCHAR(100) NOT NULL,
    culture VARCHAR(120) NOT NULL,
    unit_price NUMERIC(10,2) NOT NULL DEFAULT 0,
    cu_reference NUMERIC(6,3) NOT NULL,
    zn_reference NUMERIC(6,3) NOT NULL,
    sn_reference NUMERIC(6,3) NOT NULL,
    pb_reference NUMERIC(6,3) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_reference_alloy_services_status CHECK (status IN ('действует', 'удален'))
);

CREATE TABLE IF NOT EXISTS artifact_claims (
    id BIGSERIAL PRIMARY KEY,
    claim_code VARCHAR(40) NOT NULL UNIQUE,
    status VARCHAR(16) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    creator_id BIGINT NOT NULL,
    formed_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    moderator_id BIGINT NULL,
    artifact_title VARCHAR(180) NULL,
    artifact_origin VARCHAR(180) NULL,
    analyzer_model VARCHAR(120) NULL,
    operator_comment VARCHAR(255) NULL,
    cu_measured NUMERIC(6,3) NULL,
    zn_measured NUMERIC(6,3) NULL,
    sn_measured NUMERIC(6,3) NULL,
    pb_measured NUMERIC(6,3) NULL,
    best_match_label VARCHAR(180) NULL,
    completion_formula_result NUMERIC(8,2) NULL,
    total_cost NUMERIC(12,2) NULL,
    planned_delivery_at TIMESTAMP NULL,
    CONSTRAINT fk_claim_creator FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT fk_claim_moderator FOREIGN KEY (moderator_id) REFERENCES users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT ck_artifact_claims_status CHECK (status IN ('черновик', 'удален', 'сформирован', 'завершен', 'отклонен'))
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_claim_draft_per_creator
    ON artifact_claims (creator_id)
    WHERE status = 'черновик';

CREATE TABLE IF NOT EXISTS claim_alloy_matches (
    id BIGSERIAL PRIMARY KEY,
    claim_id BIGINT NOT NULL,
    service_id BIGINT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    sort_order INTEGER NOT NULL DEFAULT 0,
    match_value NUMERIC(10,3) NULL,
    composition_result VARCHAR(255) NULL,
    match_score NUMERIC(8,2) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_match_claim FOREIGN KEY (claim_id) REFERENCES artifact_claims(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT fk_match_service FOREIGN KEY (service_id) REFERENCES reference_alloy_services(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT ux_claim_service UNIQUE (claim_id, service_id)
);
