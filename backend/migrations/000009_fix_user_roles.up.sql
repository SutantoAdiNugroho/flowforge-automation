DO $$
BEGIN
    -- Check if 'super-admin' is already in the check constraint
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.constraint_column_usage 
        WHERE table_name = 'users' AND column_name = 'role'
        AND constraint_name = 'users_role_check'
        AND EXISTS (
            SELECT 1 FROM pg_constraint 
            WHERE conname = 'users_role_check' 
            AND pg_get_constraintdef(oid) LIKE '%super-admin%'
        )
    ) THEN
        ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
        ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('super-admin', 'admin', 'editor', 'viewer'));
    END IF;
END $$;
