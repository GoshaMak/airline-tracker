\set ON_ERROR_STOP on

SELECT format(
  'SELECT %L || %L || md5(coalesce(string_agg(row_data, E''\n'' ORDER BY row_data), '''')) FROM (SELECT row_to_json(t)::text AS row_data FROM %I.%I t) rows;',
  schemaname || '.' || tablename, ':', schemaname, tablename)
FROM pg_tables
WHERE schemaname IN ('api_schema', 'notifier_schema')
ORDER BY schemaname, tablename
\gexec
