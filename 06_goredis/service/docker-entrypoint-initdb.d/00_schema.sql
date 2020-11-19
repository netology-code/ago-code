CREATE TABLE films
(
    id          BIGSERIAL PRIMARY KEY,
    title       TEXT      NOT NULL,
    rating      FLOAT     NOT NULL,
    description TEXT      NOT NULL,
    created     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
