FROM golang:1.18-buster as builder
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 go build "-ldflags=-s -w" -trimpath -o main

FROM alpine:3.15
COPY --from=builder /go/src/app/main .
COPY --from=builder /go/src/app/static/ ./static/
ENTRYPOINT [ "./main" ]
