CREATE TABLE orbitum.audit_logs
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ  NOT NULL,
    actor_user_id   BIGINT       REFERENCES orbitum.users (id) ON DELETE SET NULL,
    actor_client_id BIGINT       REFERENCES orbitum.clients (id) ON DELETE SET NULL,
    action          VARCHAR(200) NOT NULL,
    result          VARCHAR(100),
    ip              VARCHAR(100),
    orbit_id        BIGINT       NOT NULL REFERENCES orbitum.orbits (id) ON DELETE CASCADE,
    details         JSONB
);

CREATE INDEX idx_audit_logs_action
    ON orbitum.audit_logs (action);

CREATE INDEX idx_audit_logs_contour
    ON orbitum.audit_logs (orbit_id);
