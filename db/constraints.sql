alter table passports
    add constraint number_only_digits
        check ( number ~ '^\d+$' );