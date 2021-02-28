create table wallet_operations (
    id SERIAL PRIMARY KEY,
    opertion varchar(100),
    wallet_from int,
    wallet_to int,
    amount numeric(10, 2) NOT NULL default 0.00 constraint positive_balance CHECK(amount > 0),
    created_at timestamp without time zone default current_timestamp
);