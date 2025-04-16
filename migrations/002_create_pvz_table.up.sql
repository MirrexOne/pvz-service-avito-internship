CREATE TABLE IF NOT EXISTS pvz
(
    id                UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    registration_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    city              VARCHAR(50)              NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')),
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pvz_city ON pvz (city);
CREATE INDEX IF NOT EXISTS idx_pvz_registration_date ON pvz (registration_date);

COMMENT ON TABLE pvz IS 'Таблица Пунктов Выдачи Заказов';
COMMENT ON COLUMN pvz.id IS 'Уникальный идентификатор ПВЗ (UUID)';
COMMENT ON COLUMN pvz.registration_date IS 'Дата и время регистрации ПВЗ в системе';
COMMENT ON COLUMN pvz.city IS 'Город расположения ПВЗ (Москва, Санкт-Петербург, Казань)';
COMMENT ON COLUMN pvz.created_at IS 'Время создания записи о ПВЗ';
COMMENT ON COLUMN pvz.updated_at IS 'Время последнего обновления записи о ПВЗ';
