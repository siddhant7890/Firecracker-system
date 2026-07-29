-- Item codes must be unique per admin so bulk upload can upsert by item_code.
-- Partial (item_code IS NOT NULL) since manually-added products may have none.
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_admin_item_code
    ON products(admin_id, item_code)
    WHERE item_code IS NOT NULL;
