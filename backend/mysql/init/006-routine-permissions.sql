grant execute on procedure airline_tracker.add_flight to 'dispatcher';
grant execute on procedure airline_tracker.scan_flight_info to 'dispatcher';
grant execute on procedure airline_tracker.scan_flights_info to 'dispatcher';

grant execute on procedure airline_tracker.subscribe to 'manager';
grant execute on procedure airline_tracker.scan_user_flights_info to 'manager';
