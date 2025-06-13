FROM golang:1.23 AS builder

WORKDIR /app

RUN apt update && apt install -y \
    wget \
    xz-utils \
    curl \
    && apt clean

RUN wget https://github.com/upx/upx/releases/download/v4.2.1/upx-4.2.1-amd64_linux.tar.xz
RUN tar -xJvf ./upx-4.2.1-amd64_linux.tar.xz

COPY go.mod ./
COPY go.sum ./
COPY . .

RUN go build -ldflags "-s" -o go-languageclient main.go

RUN ./upx-4.2.1-amd64_linux/upx go-languageclient

FROM node

WORKDIR /app

COPY --from=builder /app/go-languageclient .

RUN npm install -g pyright

ENTRYPOINT ["/app/go-languageclient"]
