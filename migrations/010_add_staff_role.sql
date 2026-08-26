-- Sales staff now have an explicit role: "sale_agent" (the existing mobile
-- sales-agent login flow) or "cash_agent" (cash-counter staff, who log in
-- through their own endpoint added in this change). Both roles carry
-- identical staff-level authority/access — this only separates login entry
-- points, it is not a new permission level.

ALTER TABLE sales_staff
ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'sale_agent';
