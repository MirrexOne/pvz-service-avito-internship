CREATE TABLE IF NOT EXISTS products
(
    id           UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    date_time    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    type         VARCHAR(50)              NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    reception_id UUID                     NOT NULL,
    CONSTRAINT fk_product_reception FOREIGN KEY (reception_id) REFERENCES receptions (id) ON DELETE CASCADE,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_products_reception_id ON products (reception_id);
CREATE INDEX IF NOT EXISTS idx_products_reception_datetime_desc ON products (reception_id, date_time DESC);
CREATE INDEX IF NOT EXISTS idx_products_type ON products (type);

COMMENT ON TABLE products IS 'Таблица Товаров, принятых в рамках приемок';
COMMENT ON COLUMN products.id IS 'Уникальный идентификатор товара (UUID)';
COMMENT ON COLUMN products.date_time IS 'Дата и время добавления товара в приемку';
COMMENT ON COLUMN products.type IS 'Тип товара (электроника, одежда, обувь)';
COMMENT ON COLUMN products.reception_id IS 'Внешний ключ на приемку (receptions.id)';
COMMENT ON COLUMN products.created_at IS 'Время создания записи о товаре';
COMMENT ON COLUMN products.updated_at IS 'Время последнего обновления записи о товаре';
