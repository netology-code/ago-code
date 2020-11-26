-- табличка с транзакциями
CREATE TABLE transactions
(
    id       BIGSERIAL PRIMARY KEY,
    userid   BIGINT    NOT NULL,
    category TEXT      NOT NULL,
    amount   INT      NOT NULL,
    created  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
