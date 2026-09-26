-- Admin-controlled login gate for a staff member, independent of is_active.
-- Defaults to off; only the admin update API can turn it on.

ALTER TABLE sales_staff
ADD COLUMN IF NOT EXISTS login_status BOOLEAN NOT NULL DEFAULT false;
