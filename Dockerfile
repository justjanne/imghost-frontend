FROM golang as builder

RUN curl https://glide.sh/get | sh

WORKDIR /go/src/app
COPY . .
RUN glide install
RUN CGO_ENABLED=false go build -a app .

FROM alpine:3.7
WORKDIR /root/
COPY --from=builder /go/src/app/app .
CMD ["./app"]