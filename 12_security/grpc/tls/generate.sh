#!/bin/sh

openssl req -newkey rsa:2048 -nodes -x509 -keyout key.pem -out certificate.pem \
 -subj "/C=RU/ST=Moscow/L=Moscow/O=Development/OU=Dev/CN=netology.local" \
 -addext "subjectAltName = DNS:netology.local"

