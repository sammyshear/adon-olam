FROM golang:1.24.3-alpine

WORKDIR /build

RUN apk add alsa-lib alsa-lib-dev rubberband espeak-ng sox gcc g++ pkgconfig llvm15-dev make --no-cache
ENV PATH="/opt/venv/bin:$PATH"
ENV LLVM_CONFIG="/usr/bin/llvm15-config"

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN apk add nodejs npm; npm i; npm run build; go build -o main ./cmd/main.go

EXPOSE 8080

CMD [ "/build/main" ]
