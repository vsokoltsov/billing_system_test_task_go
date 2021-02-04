create table users (
    id SERIAL PRIMARY KEY,
    email varchar(255) not null,

    unique(email)
);