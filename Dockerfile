FROM golang:1.17-buster as builder
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 go build "-ldflags=-s -w" -trimpath -o main

FROM alpine:3.15
COPY --from=builder /go/src/app/main .
COPY --from=builder /go/src/app/static/ ./static/
# "storage.SignedURL" doesn't support GCP's default credential
# so copy "cred.json" into the container and run with GOOGLE_APPLICATION_CREDENTIALS=./cred.json.
COPY --from=builder /go/src/app/cred.json .
ENTRYPOINT [ "./main" ]
