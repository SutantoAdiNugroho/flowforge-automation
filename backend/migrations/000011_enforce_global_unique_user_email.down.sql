DO $$
BEGIN
    DROP INDEX IF EXISTS idx_users_email_unique_global;

    ALTER TABLE users ADD CONSTRAINT users_tenant_id_email_key UNIQUE (tenant_id, email);

    CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
END $$;
