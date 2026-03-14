CREATE TABLE extraction_jobs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_type VARCHAR(20) NOT NULL CHECK (job_type IN ('website', 'video', 'image')),
    input_url TEXT,
    input_data BYTEA,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    error_message TEXT,
    llm_input TEXT,
    llm_output TEXT,
    recipe_id INTEGER REFERENCES recipes(id) ON DELETE SET NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_extraction_jobs_user_id ON extraction_jobs(user_id);
CREATE INDEX idx_extraction_jobs_status ON extraction_jobs(status);
CREATE INDEX idx_extraction_jobs_created_at ON extraction_jobs(created_at DESC);

CREATE TABLE extraction_feedback (
    id SERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL REFERENCES extraction_jobs(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    feedback_type VARCHAR(30) NOT NULL CHECK (feedback_type IN ('good', 'missing_info', 'inaccurate', 'other')),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(job_id, user_id)
);

CREATE INDEX idx_extraction_feedback_job_id ON extraction_feedback(job_id);
CREATE INDEX idx_extraction_feedback_rating ON extraction_feedback(rating);
