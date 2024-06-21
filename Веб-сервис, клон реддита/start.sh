#!/bin/bash

if [ ! -f docker-compose.yml ]; then
    echo "Файл docker-compose.yml не найден!"
    exit 1
fi

echo "Запускаем docker-compose..."
docker-compose up -d

sleep 5

echo "Запускаем приложение Go..."
go run ./cmd/redditclone/main.go

echo "Приложение запущено."
