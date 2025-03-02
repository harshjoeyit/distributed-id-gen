CREATE TABLE users_int (
    id INT AUTO_INCREMENT PRIMARY KEY,
    age INT
);

CREATE TABLE users_uuid (
    id VARCHAR(36) PRIMARY KEY,
    age INT
);

-- ON one of the servers execute the following statement 2 times to set ID = 2
INSERT INTO ticket (stub) VALUES ('a') ON DUPLICATE KEY UPDATE id = id + 1;