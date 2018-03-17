FROM golang:alpine as builder

RUN apk add --no-cache curl git gcc musl-dev
RUN curl https://glide.sh/get | sh

WORKDIR /go/src/app
COPY . .
RUN glide install
RUN CGO_ENABLED=false go build -a app .

FROM alpine:3.7
WORKDIR /
COPY --from=builder /go/src/app/app /app
COPY --from=builder /go/src/app/templates /templates
CMD ["/app"]