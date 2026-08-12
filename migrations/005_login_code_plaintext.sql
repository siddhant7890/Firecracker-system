-- Sales staff login codes need to be shown back to the admin in the UI and
-- can be edited later, so they can no longer be stored as a one-way hash.
-- Existing hashed rows can't be recovered as plaintext; admins must reset
-- the code for any staff created before this migration.

ALTER TABLE sales_staff RENAME COLUMN login_code_hash TO login_code;

UPDATE sales_staff SET login_code = '' WHERE login_code IS NOT NULL;
