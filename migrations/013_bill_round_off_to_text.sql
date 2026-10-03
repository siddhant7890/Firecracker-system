-- round_off is free-form text sent by the frontend, not a computed amount.

ALTER TABLE bills
ALTER COLUMN round_off TYPE TEXT USING round_off::TEXT,
ALTER COLUMN round_off SET DEFAULT '';
