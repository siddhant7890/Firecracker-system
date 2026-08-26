-- Cash Counter approve/update now also supports part-cash-part-UPI and
-- credit (pay-later) sales, alongside plain cash/upi.

ALTER TYPE payment_mode_t ADD VALUE IF NOT EXISTS 'cash_upi';
ALTER TYPE payment_mode_t ADD VALUE IF NOT EXISTS 'credit';
