FROM arm64v8/golang:1.20.4 AS builder
LABEL maintainer="Patrick Hermann patrick.hermann@sva.de"

ARG BIN="wled-resource-informer"
ARG VERSION=""
ARG BUILD_DATE=""
ARG COMMIT=""
ARG GIT_PAT=""

WORKDIR /src/
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 go build -buildvcs=false -o /bin/${BIN} \
    -ldflags="-X main.version=v${VERSION} -X main.date=${BUILD_DATE} -X main.commit=${COMMIT}"

FROM alpine:3.16.0
COPY --from=builder /bin/${BIN} /bin/${BIN}

ENTRYPOINT ["wled-resource-informer"]
