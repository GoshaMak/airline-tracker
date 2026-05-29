-- USERS
-- pswd: myStrong123
insert into users(id, email, password_hash, role)
values ('39f2f5f8-5ec3-4436-be16-341f5ef4771f', 'ab@cd.ef',
        '$2a$10$KMSc6YuqxohcaO1Zo7Cs7eSssudoNr6jzvrTCdVCWbR4ht8.9RyyK', 'user');

insert into users(id, email, password_hash, role)
values ('1caa1bd8-63f1-4cd4-adda-fb9a054407cb', 'a@b.c',
        '$2a$10$Gxo54TNRn/9qg0yoO03Csuw9asc1Htm.NuvBr3g/oWsFfL09ViQr.', 'admin');
