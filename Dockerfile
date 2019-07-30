# Intermediary Build Container
FROM golang:1.12 AS BUILDER

RUN mkdir -p /go/src/github.com/transactcharlie/scraping-spider
WORKDIR /go/src/github.com/transactcharlie/scraping-spider
COPY . .
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scraping-spider .


# Final (static from scratch container)
FROM scratch
COPY --from=BUILDER /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=BUILDER /go/src/github.com/transactcharlie/scraping-spider/scraping-spider /scraping-spider
ENTRYPOINT ["/scraping-spider"]
