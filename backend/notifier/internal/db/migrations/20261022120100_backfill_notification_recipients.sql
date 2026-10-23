-- +goose Up
-- Preserve old multi-recipient notifications as one durable delivery per address.
insert into notifier_schema.notifications (id, payload, created_at, send_at, status, type, event_id, recipient)
select gen_random_uuid(),
       jsonb_set(n.payload, '{users}', jsonb_build_array(users.user_email)),
       n.created_at, n.send_at, n.status, n.type, 'legacy-notification:' || n.id::text, users.user_email
from notifier_schema.notifications n
cross join lateral jsonb_array_elements_text(
    case when jsonb_typeof(n.payload->'users') = 'array' then n.payload->'users' else '[]'::jsonb end
) with ordinality as users(user_email, ordinal)
where n.type = 'flight_updated'
  and jsonb_array_length(case when jsonb_typeof(n.payload->'users') = 'array' then n.payload->'users' else '[]'::jsonb end) > 1
  and users.ordinal > 1
on conflict do nothing;

update notifier_schema.notifications n
set recipient = n.payload->'users'->>0,
    event_id = coalesce(n.event_id, 'legacy-notification:' || n.id::text),
    payload = jsonb_set(n.payload, '{users}', jsonb_build_array(n.payload->'users'->0))
where n.type = 'flight_updated'
  and jsonb_typeof(n.payload->'users') = 'array'
  and jsonb_array_length(case when jsonb_typeof(n.payload->'users') = 'array' then n.payload->'users' else '[]'::jsonb end) > 0;

update notifier_schema.notifications n
set recipient = n.payload->>'email',
    event_id = coalesce(n.event_id, 'legacy-notification:' || n.id::text)
where n.type = 'subscribed'
  and coalesce(n.recipient, '') = ''
  and coalesce(n.payload->>'email', '') <> '';

-- +goose Down
-- The original notification boundaries cannot be reconstructed after delivery split.
