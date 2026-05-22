CREATE TABLE tracked_accounts (
    account_id   BIGINT      PRIMARY KEY,
    display_name TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE poll_state (
    account_id     BIGINT      PRIMARY KEY,
    last_match_id  BIGINT,
    last_polled_at TIMESTAMPTZ
);

CREATE TABLE sent_notifications (
    account_id BIGINT      NOT NULL,
    match_id   BIGINT      NOT NULL,
    sent_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, match_id)
);

CREATE INDEX idx_sent_notifications_sent_at ON sent_notifications (sent_at);
