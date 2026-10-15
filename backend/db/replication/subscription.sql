SELECT format(
    'CREATE SUBSCRIPTION %I CONNECTION %L PUBLICATION airline_tracker_publication WITH (copy_data = true, synchronous_commit = ''on'', streaming = ''off'')',
    :'subscription_name', :'connection')
WHERE NOT EXISTS (SELECT FROM pg_subscription WHERE subname = :'subscription_name')
\gexec

SELECT format('ALTER SUBSCRIPTION %I CONNECTION %L', :'subscription_name', :'connection')
\gexec
SELECT format('ALTER SUBSCRIPTION %I SET (synchronous_commit = ''on'')', :'subscription_name')
\gexec
SELECT format('ALTER SUBSCRIPTION %I ENABLE', :'subscription_name')
\gexec
SELECT format('ALTER SUBSCRIPTION %I REFRESH PUBLICATION WITH (copy_data = true)', :'subscription_name')
\gexec
