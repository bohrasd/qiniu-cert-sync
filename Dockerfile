FROM golang:1.16-alpine AS builder

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64
RUN apk add make git
WORKDIR /src
COPY . .

RUN make build

FROM alpine:latest

RUN addgroup -g 1001 appgroup && \
  adduser -H -D -s /bin/false -G appgroup -u 1001 appuser

USER 1001:1001
COPY --from=builder /src/bin/qiniu-cert-sync /bin/qiniu-cert-sync
ENTRYPOINT ["/bin/qiniu-cert-sync"]
