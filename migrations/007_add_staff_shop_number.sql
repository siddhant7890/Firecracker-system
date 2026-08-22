-- Sales staff now belong to a specific shop location. Bill numbers are
-- generated per shop (not per admin) so each shop's series starts at its
-- own G-0001, e.g. SFR/G-0001/26-27 for SHOP-AKR and SFA/G-0001/26-27 for
-- SHOP-14-15.

ALTER TABLE sales_staff
ADD COLUMN IF NOT EXISTS shop_number VARCHAR(30) NOT NULL DEFAULT 'SHOP-AKR';

ALTER TABLE bill_sequences
ADD COLUMN IF NOT EXISTS shop_number VARCHAR(30) NOT NULL DEFAULT 'SHOP-AKR';

ALTER TABLE bill_sequences DROP CONSTRAINT IF EXISTS bill_sequences_pkey;
ALTER TABLE bill_sequences ADD PRIMARY KEY (admin_id, shop_number);
