FROM golang:1.15-alpine3.12 AS build
ENV CGO_ENABLED=0
COPY . /app
WORKDIR /app
RUN go build -o app.bin ./cmd
RUN chmod +x ./app.bin

FROM scratch
COPY --from=build /app/app.bin /app.bin
COPY ./keys/*.key /keys/
CMD ["/app.bin"]
