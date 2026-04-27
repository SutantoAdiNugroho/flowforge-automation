DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_versions' AND column_name='cron_expression') THEN
        ALTER TABLE workflow_versions DROP COLUMN cron_expression;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_versions' AND column_name='trigger_type') THEN
        ALTER TABLE workflow_versions DROP COLUMN trigger_type;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_versions' AND column_name='description') THEN
        ALTER TABLE workflow_versions DROP COLUMN description;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_versions' AND column_name='name') THEN
        ALTER TABLE workflow_versions DROP COLUMN name;
    END IF;
END $$;
