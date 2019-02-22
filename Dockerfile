FROM golang:1.11-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /root
COPY . .
RUN CGO_ENABLED=0 go build ./cmd/schemas 

FROM alpine:3.9
COPY --from=builder /root/schemas /root/jhu /root/scripts /

RUN chmod 700 /entrypoint.sh

CMD /entrypoint.sh

