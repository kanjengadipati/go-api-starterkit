CREATE TABLE error_messages (
  id SERIAL PRIMARY KEY,
  error_code VARCHAR(100) NOT NULL UNIQUE,
  error_type VARCHAR(50) NOT NULL,
  severity VARCHAR(20) NOT NULL,
  generic_message TEXT NOT NULL,
  ai_message TEXT,
  ai_suggestions JSONB,
  ai_generated_at TIMESTAMPTZ,
  description TEXT,
  is_sensitive BOOLEAN DEFAULT false,
  should_expose_details BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_error_code ON error_messages(error_code);
CREATE INDEX idx_error_type ON error_messages(error_type);

CREATE TABLE error_message_feedback (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  error_code VARCHAR(100) NOT NULL,
  was_helpful BOOLEAN,
  clarity_rating INTEGER,
  action_taken VARCHAR(100),
  comments TEXT,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_feedback_error_code ON error_message_feedback(error_code);
CREATE INDEX idx_feedback_user_id ON error_message_feedback(user_id);

CREATE TABLE error_analytics (
  id SERIAL PRIMARY KEY,
  error_code VARCHAR(100) NOT NULL UNIQUE,
  error_type VARCHAR(50) NOT NULL,
  occurrence_count INTEGER DEFAULT 1,
  last_occurred TIMESTAMPTZ,
  ai_message_version INTEGER DEFAULT 1,
  avg_helpfulness_rating FLOAT,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_analytics_error_code ON error_analytics(error_code);
CREATE INDEX idx_analytics_last_occurred ON error_analytics(last_occurred);

CREATE TABLE error_context_logs (
  id SERIAL PRIMARY KEY,
  user_id INTEGER,
  error_code VARCHAR(100) NOT NULL,
  error_context JSONB,
  request_path VARCHAR(255),
  request_method VARCHAR(10),
  ip_address INET,
  device_id VARCHAR(255),
  status_code INTEGER,
  response_sent JSONB,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_context_user_id ON error_context_logs(user_id);
CREATE INDEX idx_context_error_code ON error_context_logs(error_code);
CREATE INDEX idx_context_created_at ON error_context_logs(created_at);
