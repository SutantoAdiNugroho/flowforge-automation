DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_versions' AND column_name='name') THEN
        ALTER TABLE workflow_versions ADD COLUMN name VARCHAR(255);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_versions' AND column_name='description') THEN
        ALTER TABLE workflow_versions ADD COLUMN description TEXT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_versions' AND column_name='trigger_type') THEN
        ALTER TABLE workflow_versions ADD COLUMN trigger_type VARCHAR(20);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_versions' AND column_name='cron_expression') THEN
        ALTER TABLE workflow_versions ADD COLUMN cron_expression VARCHAR(100);
    END IF;

    UPDATE workflow_versions wv
    SET
        name = w.name,
        description = w.description,
        trigger_type = w.trigger_type,
        cron_expression = w.cron_expression
    FROM workflows w
    WHERE wv.workflow_id = w.id
      AND (wv.name IS NULL OR wv.trigger_type IS NULL);

    ALTER TABLE workflow_versions ALTER COLUMN name SET NOT NULL;
    ALTER TABLE workflow_versions ALTER COLUMN trigger_type SET NOT NULL;
END $$;
