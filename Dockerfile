FROM golang:1.24.3-alpine

WORKDIR /build

RUN apk add alsa-lib alsa-lib-dev python3 espeak-ng sox gcc g++ pkgconfig python3-dev llvm15-dev make --no-cache && python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"
ENV LLVM_CONFIG="/usr/bin/llvm15-config"

RUN pip install --no-cache-dir psola==0.0.1 librosa==0.11.0 --verbose

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

EXPOSE 8080

CMD [ "/build/main" ]
