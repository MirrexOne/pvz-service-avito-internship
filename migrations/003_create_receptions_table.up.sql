CREATE TABLE IF NOT EXISTS receptions
(
    id         UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    date_time  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    pvz_id     UUID                     NOT NULL,
    status     VARCHAR(20)              NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'close')),
    CONSTRAINT fk_reception_pvz FOREIGN KEY (pvz_id) REFERENCES pvz (id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX IF NOT EXISTS idx_receptions_pvz_id ON receptions (pvz_id);
CREATE INDEX IF NOT EXISTS idx_receptions_status ON receptions (status);
CREATE INDEX IF NOT EXISTS idx_receptions_pvz_status ON receptions (pvz_id, status);
CREATE INDEX IF NOT EXISTS idx_receptions_date_time ON receptions (date_time);

COMMENT ON TABLE receptions IS 'Таблица Приемок товаров на ПВЗ';
COMMENT ON COLUMN receptions.id IS 'Уникальный идентификатор приемки (UUID)';
COMMENT ON COLUMN receptions.date_time IS 'Дата и время проведения (начала) приемки';
COMMENT ON COLUMN receptions.pvz_id IS 'Внешний ключ на ПВЗ (pvz.id)';
COMMENT ON COLUMN receptions.status IS 'Статус приемки (in_progress, close)';
COMMENT ON COLUMN receptions.created_at IS 'Время создания записи о приемке';
COMMENT ON COLUMN receptions.updated_at IS 'Время последнего обновления записи о приемке';
