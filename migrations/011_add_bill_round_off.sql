-- Bill-level round-off applied on the "New Bill" screen.

ALTER TABLE bills
ADD COLUMN IF NOT EXISTS round_off NUMERIC(12,2) NOT NULL DEFAULT 0;
