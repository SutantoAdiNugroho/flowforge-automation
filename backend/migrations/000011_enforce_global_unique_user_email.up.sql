DO $$
BEGIN
    ALTER TABLE users DROP CONSTRAINT IF EXISTS users_tenant_id_email_key;

    DROP INDEX IF EXISTS idx_users_email;

    CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique_global ON users (LOWER(email));
END $$;
