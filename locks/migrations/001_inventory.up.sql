CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    balance INT
);

INSERT INTO accounts (id, name, balance) VALUES (1, 'Alice', 1000);