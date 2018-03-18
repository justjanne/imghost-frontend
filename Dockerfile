FROM golang:alpine as go_builder

RUN apk add --no-cache curl git gcc musl-dev
RUN curl https://glide.sh/get | sh

WORKDIR /go/src/app
COPY glide.* ./
RUN glide install
COPY *.go ./
RUN CGO_ENABLED=false go build -a app .

FROM node:alpine as asset_builder
WORKDIR /app
COPY package* /app/
COPY assets /app/assets
RUN npm install
RUN npm run build

FROM alpine:3.7
WORKDIR /
COPY --from=go_builder /go/src/app/app /app
COPY templates /templates
COPY --from=asset_builder /app/assets /assets
ENTRYPOINT ["/app"]