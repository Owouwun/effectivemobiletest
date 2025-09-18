CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS services (
    service_name TEXT PRIMARY KEY,
    price INT NOT NULL,
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE
);

CREATE INDEX idx_services_name ON services(service_name);
CREATE INDEX idx_services_price ON services(price);
CREATE INDEX idx_services_user_id ON services(user_id);