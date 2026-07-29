-- SalesTrack initial schema
-- Run with: psql "$DATABASE_URL" -f migrations/001_init.sql

DO $$ BEGIN
    CREATE TYPE bill_status AS ENUM ('pending', 'approved', 'rejected');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE payment_mode_t AS ENUM ('cash', 'upi');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Shop owner / admin portal login
CREATE TABLE IF NOT EXISTS admins (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(120) NOT NULL,
    mobile_number   VARCHAR(15) UNIQUE NOT NULL,
    email           VARCHAR(150) UNIQUE,
    password_hash   TEXT NOT NULL,
    shop_name       VARCHAR(150) NOT NULL DEFAULT 'My Shop',
    bill_prefix     VARCHAR(10) NOT NULL DEFAULT 'SF',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Sales agents (mobile app users), created by admin from User Management screen
CREATE TABLE IF NOT EXISTS sales_staff (
    id              SERIAL PRIMARY KEY,
    admin_id        INT NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    name            VARCHAR(120) NOT NULL,
    mobile_number   VARCHAR(15) UNIQUE NOT NULL,
    login_code_hash TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Per-admin running bill number sequence (used to build bill_no like SF/26-27/00483)
CREATE TABLE IF NOT EXISTS bill_sequences (
    admin_id    INT PRIMARY KEY REFERENCES admins(id) ON DELETE CASCADE,
    last_number INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS products (
    id              SERIAL PRIMARY KEY,
    admin_id        INT NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    name            VARCHAR(150) NOT NULL,
    category        VARCHAR(100) NOT NULL,
    hsn_code        VARCHAR(20) NOT NULL DEFAULT '3604',
    unit            VARCHAR(30) NOT NULL DEFAULT 'Piece',
    taxable_value   NUMERIC(12,2) NOT NULL,
    gst_percent     NUMERIC(5,2) NOT NULL,
    final_price     NUMERIC(12,2) GENERATED ALWAYS AS (ROUND(taxable_value * (1 + gst_percent / 100.0), 2)) STORED,
    image_url       TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_products_admin ON products(admin_id, is_active);

CREATE TABLE IF NOT EXISTS bills (
    id                  SERIAL PRIMARY KEY,
    admin_id            INT NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    sales_staff_id      INT NOT NULL REFERENCES sales_staff(id),
    bill_no             VARCHAR(40) NOT NULL,
    financial_year      VARCHAR(10) NOT NULL,
    customer_name       VARCHAR(150) NOT NULL DEFAULT 'Walk-in Customer',
    customer_mobile     VARCHAR(15),
    taxable_amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
    cgst_amount         NUMERIC(12,2) NOT NULL DEFAULT 0,
    sgst_amount         NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_amount        NUMERIC(12,2) NOT NULL DEFAULT 0,
    status              bill_status NOT NULL DEFAULT 'pending',
    payment_mode        payment_mode_t,
    razorpay_order_id   VARCHAR(100),
    razorpay_payment_id VARCHAR(100),
    whatsapp_sent       BOOLEAN NOT NULL DEFAULT false,
    whatsapp_sent_at    TIMESTAMPTZ,
    approved_by         INT REFERENCES admins(id),
    approved_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (admin_id, bill_no)
);
CREATE INDEX IF NOT EXISTS idx_bills_admin_created ON bills(admin_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bills_staff ON bills(sales_staff_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bills_status ON bills(admin_id, status);

CREATE TABLE IF NOT EXISTS bill_items (
    id              SERIAL PRIMARY KEY,
    bill_id         INT NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    product_id      INT REFERENCES products(id),
    product_name    VARCHAR(150) NOT NULL,
    hsn_code        VARCHAR(20) NOT NULL,
    qty             INT NOT NULL,
    rate            NUMERIC(12,2) NOT NULL,
    taxable_amount  NUMERIC(12,2) NOT NULL,
    gst_percent     NUMERIC(5,2) NOT NULL,
    cgst_percent    NUMERIC(5,2) NOT NULL,
    cgst_amount     NUMERIC(12,2) NOT NULL,
    sgst_percent    NUMERIC(5,2) NOT NULL,
    sgst_amount     NUMERIC(12,2) NOT NULL,
    total_amount    NUMERIC(12,2) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_bill_items_bill ON bill_items(bill_id);
