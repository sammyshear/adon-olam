FROM golang:1.24.3-alpine AS buildgo

WORKDIR /build

RUN apk add alsa-lib alsa-lib-dev rubberband espeak-ng sox gcc g++ pkgconfig llvm15-dev make --no-cache
ENV LLVM_CONFIG="/usr/bin/llvm15-config"

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go tool templ generate; go build -o main ./cmd/main.go

FROM oven/bun:alpine AS buildbun

WORKDIR /build

RUN apk add alsa-lib alsa-lib-dev rubberband espeak-ng sox gcc g++ pkgconfig llvm15-dev make --no-cache
ENV LLVM_CONFIG="/usr/bin/llvm15-config"

COPY . .

RUN bun i; bun run build;

EXPOSE 8080

COPY --from=buildgo /build/main /build/main

CMD [ "/build/main" ]
