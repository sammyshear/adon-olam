FROM oven/bun:alpine AS buildbun

WORKDIR /buildbun

COPY . .

RUN bun i; bun run build;

FROM golang:1.24.3-alpine AS buildgo

WORKDIR /build
RUN apk add alsa-lib alsa-lib-dev rubberband espeak-ng sox gcc g++ pkgconfig llvm15-dev make --no-cache
ENV LLVM_CONFIG="/usr/bin/llvm15-config"

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY --from=buildbun /buildbun/static ./static

RUN go tool templ generate; go build -o main ./cmd/main.go


EXPOSE 8080
CMD [ "/build/main" ]
