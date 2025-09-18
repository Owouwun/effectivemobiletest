DROP INDEX IF EXISTS idx_services_user_id;
DROP INDEX IF EXISTS idx_services_price;
DROP INDEX IF EXISTS idx_services_name;
DROP TABLE IF EXISTS services;
DROP EXTENSION IF EXISTS "uuid-ossp";