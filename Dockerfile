FROM debian:latest

# 复制二进制和data目录到/app
COPY honeypot /app/honeypot
COPY data /app/data
COPY templates /app/templates

RUN chmod +x /app/honeypot

WORKDIR /app

CMD ["./honeypot"]

