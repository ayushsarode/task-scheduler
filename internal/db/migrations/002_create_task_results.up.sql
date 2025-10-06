CREATE TABLE IF NOT EXISTS task_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    run_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status_code INTEGER NOT NULL DEFAULT 0,
    success BOOLEAN NOT NULL DEFAULT false,
    response_headers JSONB,
    response_body TEXT,
    error_message TEXT,
    duration_ms BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_results_task_id ON task_results(task_id);
CREATE INDEX idx_task_results_run_at ON task_results(run_at);
CREATE INDEX idx_task_results_success ON task_results(success);
CREATE INDEX idx_task_results_created_at ON task_results(created_at);