-- +goose Up
create table users (
    id serial primary key,
    name varchar(30) not null,
    phone varchar(30) not null
)

create table user_facility_limits {
    id serial primary key,
    user_id int not null references users(id),
    limit_amount decimal(15,2) not null
}

create table tenors {
    id serial primary key,
    tenor_value int not null unique
}

create table user_facilities (
    id serial primary key,
    user_id int not null references users(id),
    facility_limit_id int not null references user_facility_limits(id),
    amount decimal(15,2) not null,
    tenor int not null,
    monthly_installment decimal(15,2) not null,
    total_margin decimal(15,2) not null,
    total_payment decimal(15,2) not null,
    created_at timestamp default current_timestamp
)

create table user_facility_details (
    id serial primary key,
    user_facility_id int not null references user_facilities(id),
    due_date date not null,
    installment_amount decimal(15,2) not null
)

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
drop table users
drop table user_facility_limits
drop table tenors
drop table user_facilities
drop table user_facility_details
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
