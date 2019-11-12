FROM golang:alpine as go_builder

RUN apk add --no-cache musl-dev

WORKDIR /go/src/app
COPY *.go go.* ./
RUN go mod download
RUN CGO_ENABLED=false go build -o app .

FROM node:alpine as asset_builder
WORKDIR /app
COPY package* /app/
RUN npm install
COPY assets /app/assets
RUN npm run build

FROM alpine:3.10
WORKDIR /
COPY --from=go_builder /go/src/app/app /app
COPY templates /templates
COPY --from=asset_builder /app/assets /assets
ENTRYPOINT ["/app"]