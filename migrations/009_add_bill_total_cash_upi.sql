-- Split totals for the "cash_upi" payment mode (part cash, part UPI),
-- editable via the bill update endpoint.

ALTER TABLE bills
ADD COLUMN IF NOT EXISTS total_cash NUMERIC(12,2),
ADD COLUMN IF NOT EXISTS total_upi  NUMERIC(12,2);
