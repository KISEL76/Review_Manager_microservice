CREATE TABLE pr_statuses (
    id   INT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

INSERT INTO pr_statuses (id, name) VALUES
    (1, 'OPEN'),
    (2, 'MERGED');

CREATE TABLE teams (
    id        SERIAL PRIMARY KEY,
    team_name TEXT NOT NULL UNIQUE
);

CREATE TABLE users (
    user_id   TEXT PRIMARY KEY,
    username  TEXT NOT NULL,
    team_id   INT NOT NULL REFERENCES teams (id) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Индекс под выборку кандидатов в ревьюеры:
CREATE INDEX idx_users_team_active ON users (team_id, is_active);

CREATE TABLE pull_requests (
    pull_request_id   TEXT PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id         TEXT NOT NULL REFERENCES users (user_id),
    status_id         INT NOT NULL REFERENCES pr_statuses (id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at         TIMESTAMPTZ,
    
    need_more_reviewers BOOLEAN NOT NULL DEFAULT FALSE
);

-- Индекс для запросов/статистики по автору:
CREATE INDEX idx_pull_requests_author ON pull_requests (author_id);

-- Индекс по статусу PR:
CREATE INDEX idx_pull_requests_status ON pull_requests (status_id);

CREATE TABLE pull_request_reviewers (
    pull_request_id TEXT NOT NULL REFERENCES pull_requests (pull_request_id) ON DELETE CASCADE,
    reviewer_id     TEXT NOT NULL REFERENCES users (user_id),
    PRIMARY KEY (pull_request_id, reviewer_id)
);

-- Индекс под /users/getReview:
CREATE INDEX idx_pr_reviewers_reviewer_id ON pull_request_reviewers (reviewer_id);