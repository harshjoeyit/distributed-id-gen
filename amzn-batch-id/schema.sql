USE test;

CREATE TABLE amazon_id (
    service_name VARCHAR(255) PRIMARY KEY, 
    counter bigint
);

INSERT INTO amazon_id (service_name) VALUES ('order'), ('product');