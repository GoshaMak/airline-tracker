DO $$
DECLARE relation record; has_rows boolean;
BEGIN
    FOR relation IN SELECT schemaname, tablename FROM pg_tables
        WHERE schemaname IN ('api_schema', 'notifier_schema', 'replication_meta')
            AND tablename <> 'goose_db_version'
    LOOP
        EXECUTE format('SELECT EXISTS (SELECT FROM %I.%I LIMIT 1)',
            relation.schemaname, relation.tablename) INTO has_rows;
        IF has_rows THEN
            RAISE EXCEPTION 'Subscriber contains data but has no expected subscription; refusing initial copy';
        END IF;
    END LOOP;
END $$;
