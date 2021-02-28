create table wallets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    balance numeric(10, 2) NOT NULL default 0.00 constraint positive_balance CHECK(balance > 0),
    currency varchar(5) NOT NULL default 'USD',
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id)
);