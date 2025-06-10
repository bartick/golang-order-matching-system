-- Setup configurations
SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;
COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(10) NOT NULL,
    side VARCHAR(4) NOT NULL CHECK (side IN ('buy', 'sell')),
    type VARCHAR(6) NOT NULL CHECK (type IN ('limit', 'market')),
    price DECIMAL(10, 2) NULL, -- NULL for market orders
    initial_quantity INTEGER NOT NULL,
    remaining_quantity INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'filled', 'canceled', 'partially_filled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_quantity_positive CHECK (initial_quantity > 0),
    CONSTRAINT chk_remaining_quantity CHECK (remaining_quantity >= 0),
    CONSTRAINT chk_remaining_lte_initial CHECK (remaining_quantity <= initial_quantity),
    CONSTRAINT chk_limit_order_has_price CHECK (type = 'market' OR price IS NOT NULL),
    CONSTRAINT chk_price_positive CHECK (price IS NULL OR price > 0)
);

CREATE TABLE trades (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    buy_order_id UUID NOT NULL,
    sell_order_id UUID NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    quantity INTEGER NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    FOREIGN KEY (buy_order_id) REFERENCES orders(id),
    FOREIGN KEY (sell_order_id) REFERENCES orders(id),
    
    -- Constraints
    CONSTRAINT chk_trade_quantity_positive CHECK (quantity > 0),
    CONSTRAINT chk_trade_price_positive CHECK (price > 0)
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_orders_updated_at 
    BEFORE UPDATE ON orders 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX idx_orders_symbol_side_status ON orders(symbol, side, status);
CREATE INDEX idx_orders_symbol_side_price_created ON orders(symbol, side, price, created_at);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);

CREATE INDEX idx_buy_orders ON orders(symbol, side, price DESC, created_at ASC) 
WHERE side = 'buy' AND status IN ('open', 'partially_filled');

CREATE INDEX idx_sell_orders ON orders(symbol, side, price ASC, created_at ASC)
WHERE side = 'sell' AND status IN ('open', 'partially_filled');

CREATE INDEX idx_trades_symbol ON trades(symbol);
CREATE INDEX idx_trades_executed_at ON trades(executed_at);
CREATE INDEX idx_trades_buy_order_id ON trades(buy_order_id);
CREATE INDEX idx_trades_sell_order_id ON trades(sell_order_id);

CREATE TABLE order_book_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(10) NOT NULL,
    side VARCHAR(4) NOT NULL CHECK (side IN ('buy', 'sell')),
    price DECIMAL(10, 2) NOT NULL,
    total_quantity INTEGER NOT NULL,
    order_count INTEGER NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE (symbol, side, price)
);

-- Sample data for testing
INSERT INTO orders (symbol, side, type, price, initial_quantity, remaining_quantity, status) VALUES
-- Buy orders for AAPL
('AAPL', 'buy', 'limit', 150.00, 100, 100, 'open'),
('AAPL', 'buy', 'limit', 149.50, 200, 200, 'open'),
('AAPL', 'buy', 'limit', 149.00, 150, 150, 'open'),

-- Sell orders for AAPL
('AAPL', 'sell', 'limit', 151.00, 100, 100, 'open'),
('AAPL', 'sell', 'limit', 151.50, 200, 200, 'open'),
('AAPL', 'sell', 'limit', 152.00, 150, 150, 'open'),

-- Buy orders for GOOGL
('GOOGL', 'buy', 'limit', 2800.00, 50, 50, 'open'),
('GOOGL', 'buy', 'limit', 2795.00, 25, 25, 'open'),

-- Sell orders for GOOGL
('GOOGL', 'sell', 'limit', 2805.00, 30, 30, 'open'),
('GOOGL', 'sell', 'limit', 2810.00, 40, 40, 'open');


CREATE VIEW active_orders AS
SELECT * FROM orders 
WHERE status IN ('open', 'partially_filled')
ORDER BY symbol, side, 
    CASE WHEN side = 'buy' THEN price END DESC,
    CASE WHEN side = 'sell' THEN price END ASC,
    created_at ASC;

CREATE VIEW order_book AS
SELECT 
    symbol,
    side,
    price,
    SUM(remaining_quantity) as total_quantity,
    COUNT(*) as order_count,
    MIN(created_at) as earliest_order
FROM orders 
WHERE status IN ('open', 'partially_filled')
GROUP BY symbol, side, price
ORDER BY symbol, side,
    CASE WHEN side = 'buy' THEN price END DESC,
    CASE WHEN side = 'sell' THEN price END ASC;

CREATE VIEW recent_trades AS
SELECT 
    t.*,
    bo.symbol as buy_symbol,
    so.symbol as sell_symbol
FROM trades t
JOIN orders bo ON t.buy_order_id = bo.id
JOIN orders so ON t.sell_order_id = so.id
ORDER BY t.executed_at DESC;