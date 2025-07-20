FROM oven/bun:alpine AS buildbun

WORKDIR /buildbun

COPY . .

RUN bun i; bun run build;

FROM golang:1.24.3-alpine AS buildgo

WORKDIR /build
RUN apk add alsa-lib alsa-lib-dev espeak-ng sox gcc g++ pkgconfig llvm15-dev make --no-cache
ENV LLVM_CONFIG="/usr/bin/llvm15-config"

RUN wget -q https://github.com/praat/praat.github.io/releases/download/v6.4.39/praat6439_linux-intel64-barren.tar.gz; tar xvf praat6439_linux-intel64-barren.tar.gz
RUN mv ./praat_barren /usr/local/bin/praat

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY --from=buildbun /buildbun/static ./static

RUN go tool templ generate; go build -o main ./cmd/main.go


EXPOSE 8080
CMD [ "/build/main" ]
