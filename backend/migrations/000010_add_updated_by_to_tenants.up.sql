DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='tenants' AND column_name='updated_by') THEN
        ALTER TABLE tenants ADD COLUMN updated_by UUID;
    END IF;
END $$;
