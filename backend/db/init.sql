-- Database Schema for Stage 1: Materials and Products

CREATE TABLE IF NOT EXISTS materials (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    unit VARCHAR(20) NOT NULL,
    on_hand DECIMAL(15,3) NOT NULL DEFAULT 0.000,
    reserved DECIMAL(15,3) NOT NULL DEFAULT 0.000,
    version INT UNSIGNED NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS products (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    unit VARCHAR(20) NOT NULL,
    on_hand DECIMAL(15,3) NOT NULL DEFAULT 0.000,
    reserved DECIMAL(15,3) NOT NULL DEFAULT 0.000,
    version INT UNSIGNED NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed Data

INSERT INTO materials (sku, name, unit, on_hand, reserved) VALUES
('RM-001', 'Kain', 'gram', 10000.000, 0.000),
('RM-002', 'Benang', 'gram', 5000.000, 0.000),
('RM-003', 'Kancing', 'pcs', 1000.000, 0.000)
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    unit = VALUES(unit),
    on_hand = VALUES(on_hand),
    reserved = VALUES(reserved);

INSERT INTO products (sku, name, unit, on_hand, reserved) VALUES
('FG-001', 'Kemeja', 'pcs', 10.000, 0.000)
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    unit = VALUES(unit),
    on_hand = VALUES(on_hand),
    reserved = VALUES(reserved);
