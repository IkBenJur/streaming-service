#!/bin/sh
set -e
envsubst '${BACKEND_URL}' < /usr/share/nginx/html/index.html.template > /usr/share/nginx/html/index.html
exec nginx -g 'daemon off;'
