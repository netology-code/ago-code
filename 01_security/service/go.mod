module service

go 1.14

require (
	github.com/google/uuid v1.1.1
	github.com/jackc/pgx/v4 v4.8.1
	github.com/netology-code/remux v0.0.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
)

// Инструкция replace позволяет вам не скачивать каждый раз с GitHub/etc, а просто ссылаться на указанный каталог локально
replace github.com/netology-code/remux => ../remux
