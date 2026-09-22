-- Database Schema for Materials, Products, and Bill of Materials (BOM)

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

CREATE TABLE IF NOT EXISTS boms (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT UNSIGNED NOT NULL,
    version INT UNSIGNED NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_boms_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    UNIQUE KEY uq_product_version (product_id, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS bom_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    bom_id BIGINT UNSIGNED NOT NULL,
    material_id BIGINT UNSIGNED NOT NULL,
    quantity DECIMAL(15,3) NOT NULL,
    CONSTRAINT fk_bom_items_bom FOREIGN KEY (bom_id) REFERENCES boms(id) ON DELETE CASCADE,
    CONSTRAINT fk_bom_items_material FOREIGN KEY (material_id) REFERENCES materials(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed Data

INSERT INTO materials (id, sku, name, unit, on_hand, reserved) VALUES
(1, 'RM-001', 'Kain', 'gram', 10000.000, 0.000),
(2, 'RM-002', 'Benang', 'gram', 5000.000, 0.000),
(3, 'RM-003', 'Kancing', 'pcs', 1000.000, 0.000)
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    unit = VALUES(unit),
    on_hand = VALUES(on_hand),
    reserved = VALUES(reserved);

INSERT INTO products (id, sku, name, unit, on_hand, reserved) VALUES
(1, 'FG-001', 'Kemeja', 'pcs', 10.000, 0.000)
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    unit = VALUES(unit),
    on_hand = VALUES(on_hand),
    reserved = VALUES(reserved);

-- Seed BOM for FG-001 version 1
INSERT INTO boms (id, product_id, version, is_active) VALUES
(1, 1, 1, TRUE)
ON DUPLICATE KEY UPDATE is_active = VALUES(is_active);

INSERT INTO bom_items (bom_id, material_id, quantity) VALUES
(1, 1, 500.000),
(1, 2, 50.000),
(1, 3, 5.000)
ON DUPLICATE KEY UPDATE quantity = VALUES(quantity);
