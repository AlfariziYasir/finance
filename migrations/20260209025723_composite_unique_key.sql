-- +goose Up
alter table users
add constraint unique_user unique (phone, name);

alter table user_facility_limits
add constraint unique_user_limit unique (user_id);

alter table user_facilities
add constraint unique_user_facilities unique (user_id, facility_limit_id, start_date, amount, tenor, created_at);

alter table user_facility_details
add constraint unique_user_facility_details unique (user_facility_id, due_date);


-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
alter table user_facility_details drop constraint if exists unique_user_facility_details;
alter table user_facilities drop constraint if exists unique_user_facilities;
alter table user_facility_limits drop constraint if exists unique_user_limit;
alter table users drop constraint if exists unique_user;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
