1. docker-compose up
2. Перейти по адресу localhost:8500, дождаться пока сервисы auth и transactions зарегистрируются
3. Сделать по очереди запросы из requests.http (показать, что всё работает)
4. Сделать docker-compose stop transactions
5. Показать интерфейс localhost:8500
6. Сделать запрос на получение транзакции
7. Поднять сервис транзакций (можно на другом IP): docker-compose up transactions
8. Показать, что запросы проходят
9. Открыть localhost:16686, показать distributed tracing