-- +goose Up
-- +goose StatementBegin
create or replace function subscribe(p_uid uuid, p_fid uuid)
    returns void
    language plpgsql
as
$$
begin
    if not exists(select 1 from users where id = p_uid) then
        raise exception 'user not found' using errcode = 'P0002';
    end if;

    if not exists(select 1 from flights where id = p_fid) then
        raise exception 'flight not found' using errcode = 'P0002';
    end if;

    insert into subscriptions(user_id, flight_id)
    values (p_uid, p_fid);
end;
$$;
-- +goose StatementEnd

-- +goose Down
drop function if exists subscribe(p_uid uuid, p_fid uuid);