-- Sales staff can now approve bills from the Cash Counter too, so
-- bills.approved_by is no longer always an admins(id): drop the FK and
-- record which table the id belongs to alongside it.

ALTER TABLE bills DROP CONSTRAINT IF EXISTS bills_approved_by_fkey;

ALTER TABLE bills
ADD COLUMN IF NOT EXISTS approved_by_role VARCHAR(20);

-- Everything approved before this migration was actioned by an admin.
UPDATE bills SET approved_by_role = 'admin' WHERE approved_by IS NOT NULL AND approved_by_role IS NULL;
