-- +goose Up
insert into users (id, name, phone) 
values (1, 'Khabib Nurmagomedov', '08123456789'),
    (2, 'Islam Makhachev', '08987654321'),
    (3, 'Khamzat Chimaev', '08234567891');

insert into user_user_facility_limits (id, user_id, limit_amount)
values (1, 1, 10000000),
    (2, 2, 15000000),
    (3, 3, 20000000),

insert into tenors (id, tenor_value) values (1, 6), (2, 12), (3, 18), (4, 24), (5, 30), (6, 36);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
truncate table users;
truncate table tenors;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
