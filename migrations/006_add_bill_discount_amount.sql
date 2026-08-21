-- Bill-level discount applied on the "New Bill" screen.

ALTER TABLE bills
ADD COLUMN IF NOT EXISTS discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0;
