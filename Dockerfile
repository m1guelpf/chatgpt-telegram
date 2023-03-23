FROM golang:1-alpine as builder
RUN apk update && apk add gcc make g++ git
WORKDIR /build
ADD . .
RUN make build

FROM alpine
COPY --from=builder /build/chatgpt-telegram /bin/chatgpt-telegram
RUN chmod +x /bin/chatgpt-telegram && mkdir -p /root/.config

ENV TELEGRAM_ID ""
ENV TELEGRAM_TOKEN ""
ENV OPENAI_SESSION ""

ENTRYPOINT ["/bin/chatgpt-telegram"]