FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o /server ./cmd/server

FROM alpine:latest
COPY --from=build /server /server
EXPOSE 9000
CMD ["/server"]
