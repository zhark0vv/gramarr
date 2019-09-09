FROM golang:1.13-alpine AS build
RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o gramarr  ./cmd/app

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /app/gramarr .

CMD [ "/root/gramarr", "--configDir", "/config" ]