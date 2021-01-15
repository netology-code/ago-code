FROM golang:1.15-alpine3.12 AS build
ENV CGO_ENABLED=0
COPY . /app
WORKDIR /app
RUN go build -o app.bin ./cmd
RUN chmod +x ./app.bin

FROM alpine:3.12
RUN apk add postgresql-client
COPY --from=build /app/app.bin /app.bin
COPY ./wait-for-postgres.sh /wait-for-postgres.sh
RUN chmod +x ./wait-for-postgres.sh
COPY ./keys/*.key /keys/
CMD ["/wait-for-postgres.sh", "/app.bin"]
