FROM alpine:3.7

# add bash and openssh
RUN apk add --update --no-cache bash netcat-openbsd

# simple http response
COPY ./ok.http /ok.http
COPY ./start.sh /start.sh

WORKDIR /

# netcat will listen for incomming traffic on this port and return the ok response
CMD ["sh", "start.sh"]