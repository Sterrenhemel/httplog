FROM alpine

COPY deploy /app

ENV ENV="dev"
ENV SERVICE_NAME=httplog

WORKDIR /app
CMD ["sh", "./bootstrap.sh"]