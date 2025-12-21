CREATE TABLE IF NOT EXISTS subscriptions(
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    start_date DATE  NOT NULL,
    end_date DATE
);

CREATE INDEX IF NOT EXISTS idx_user_service
    ON subscriptions(user_id, service_name);

CREATE INDEX IF NOT EXISTS idx_subscriptions_period
    ON subscriptions(start_date, end_date);