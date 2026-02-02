-- 1. Enable UUID Extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 2. Auth & RBAC
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL, -- Admin, StoreManager, StitchingManager
    permissions TEXT -- JSON blob
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role_id INT REFERENCES roles(id),
    branch_id INT, -- Nullable (Admin has no branch)
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Locations (Store & Factory)
CREATE TABLE branches (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address TEXT,
    manager_id UUID REFERENCES users(id)
);

CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    branch_id INT REFERENCES branches(id),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) DEFAULT 'STORE' -- CENTRAL, STORE, FACTORY
);

-- 4. Product Catalog (The "Meters" Logic)
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category_id INT REFERENCES categories(id),
    brand VARCHAR(100),
    is_web_visible BOOLEAN DEFAULT TRUE,  -- Hides "Cloth Sheets" from Web
    is_stitched BOOLEAN DEFAULT FALSE,    -- True = Shirt (We made it)
    uom VARCHAR(20) DEFAULT 'Unit'        -- 'Unit' or 'Meter'
);

CREATE TABLE variants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID REFERENCES products(id),
    name VARCHAR(100),       -- "Size M - Red"
    sku VARCHAR(50) UNIQUE NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    cost_price DECIMAL(10, 2)
);

CREATE TABLE barcodes (
    id SERIAL PRIMARY KEY,
    variant_id UUID REFERENCES variants(id),
    code VARCHAR(100) UNIQUE NOT NULL,
    generated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Inventory & Suppliers
CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    contact VARCHAR(100),
    email VARCHAR(150)
);

CREATE TABLE warehouse_stocks (
    id SERIAL PRIMARY KEY,
    warehouse_id INT REFERENCES warehouses(id),
    variant_id UUID REFERENCES variants(id),
    quantity DECIMAL(10, 2) DEFAULT 0,    -- Supports 1.50 meters
    reserved DECIMAL(10, 2) DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(warehouse_id, variant_id)      -- Prevents duplicate rows
);

CREATE TABLE stock_movements (
    id SERIAL PRIMARY KEY,
    variant_id UUID REFERENCES variants(id),
    from_warehouse_id INT REFERENCES warehouses(id),
    to_warehouse_id INT REFERENCES warehouses(id),
    quantity DECIMAL(10, 2),
    type VARCHAR(50),                     -- PURCHASE, SALE, TRANSFER, MANUFACTURING
    reference_id VARCHAR(100),            -- Invoice ID or PO Number
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 6. Billing & Finance (Split Payments)
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    branch_id INT REFERENCES branches(id),
    total_amount DECIMAL(10, 2) NOT NULL,
    tax_amount DECIMAL(10, 2) DEFAULT 0,
    type VARCHAR(20) DEFAULT 'POS',       -- POS or WEB
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    invoice_id UUID REFERENCES invoices(id),
    method VARCHAR(50),                   -- Cash, Card, UPI, StoreCredit
    amount DECIMAL(10, 2) NOT NULL,
    reference VARCHAR(100)                -- Trans ID
);