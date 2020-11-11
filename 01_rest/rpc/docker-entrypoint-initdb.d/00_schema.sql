CREATE TABLE offers
(
    id      BIGSERIAL PRIMARY KEY,
    company TEXT      NOT NULL,
    percent TEXT      NOT NULL,
    comment TEXT      NOT NULL,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
