FROM golang:1.19 as build

    ARG prototype

    WORKDIR /go/src/app
    COPY $prototype .

    RUN go mod download
    RUN go vet -v

    RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian11
    USER 1000:1000
    VOLUME "/data"

    COPY --from=build /go/bin/app /

    ENTRYPOINT [ "/app" ]
    CMD [ "server",  "0.0.0.0:9000" ]
