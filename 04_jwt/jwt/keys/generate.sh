#!/bin/sh

openssl rand -out symmetric.key -base64 256 # base64 256 бит
openssl genrsa -out private.key 2048 # приватный ключ 2048 бит
openssl rsa -pubout -in private.key -out public.key # публичный ключ из приватного
