CREATE TABLE IF NOT EXISTS users
(
    id            UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    email         VARCHAR(255)             NOT NULL UNIQUE,
    password_hash VARCHAR(255)             NOT NULL,
    role          VARCHAR(20)              NOT NULL CHECK (role IN ('employee', 'moderator')),
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

COMMENT ON TABLE users IS 'Таблица для хранения информации о пользователях системы ПВЗ';
COMMENT ON COLUMN users.id IS 'Уникальный идентификатор пользователя (UUID)';
COMMENT ON COLUMN users.email IS 'Электронная почта пользователя (уникальная)';
COMMENT ON COLUMN users.password_hash IS 'Хеш пароля пользователя (bcrypt)';
COMMENT ON COLUMN users.role IS 'Роль пользователя (employee, moderator)';
COMMENT ON COLUMN users.created_at IS 'Время создания записи о пользователе';
COMMENT ON COLUMN users.updated_at IS 'Время последнего обновления записи о пользователе';
