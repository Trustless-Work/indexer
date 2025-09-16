-- +goose Up
-- Pega aquí TODO el contenido del plain_db_2.txt (tal cual).
-- No elimines los CREATE FUNCTION / CREATE TABLE / INDEX / TRIGGER, etc.

-- +goose Down
-- ⚠️ Solo para DEV: borra todo y recrea schema limpio.
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
