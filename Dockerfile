FROM golang:1.18 as builder

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 go build -o /bin/markdown-enricher

FROM gcr.io/distroless/static as runtime

COPY --from=builder /bin/markdown-enricher /

CMD ["/markdown-enricher"]
