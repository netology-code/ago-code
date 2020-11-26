-- табличка с пользователями и их паролями
CREATE TABLE users
(
    id       BIGSERIAL PRIMARY KEY,
    login    TEXT      NOT NULL UNIQUE,
    password TEXT      NOT NULL,
    created  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- табличка с токенами (чтобы пользователь каждый раз не присылал пароль и логин)
CREATE TABLE tokens (
    token TEXT PRIMARY KEY,
    userid BIGINT NOT NULL REFERENCES users,
    created  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);