-- Extra bill-level fields captured on the "New Bill" screen.

ALTER TABLE bills
ADD COLUMN IF NOT EXISTS token_number     VARCHAR(40),
ADD COLUMN IF NOT EXISTS number_of_cartoon INT,
ADD COLUMN IF NOT EXISTS gst_number       VARCHAR(20);
