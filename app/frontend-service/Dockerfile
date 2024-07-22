FROM nginx:stable-alpine-slim

COPY nginx.conf /etc/nginx/conf.d/default.conf

COPY page/. /usr/share/nginx/html

EXPOSE 80