FROM caddy:alpine

ARG VERSION=latest

COPY dist/ /usr/share/caddy/

COPY Caddyfile /etc/caddy/Caddyfile

RUN echo $VERSION > /usr/share/caddy/version.txt

EXPOSE 80 443
