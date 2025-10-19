#!/bin/bash

# Функция для обработки прерывания
cleanup() {
    echo "Завершение работы..."
    # Останавливаем все возможные контейнеры
    docker compose --env-file quarkus.env --env-file .env down
    docker compose --env-file micronaut.env --env-file .env down
    docker compose --env-file spring.env --env-file .env down
    exit 0
}

# Устанавливаем обработчик сигналов
trap cleanup SIGINT SIGTERM

echo "Последовательный запуск docker-compose с разными env-файлами"
echo "Каждая конфигурация работает 30 секунд в фоновом режиме"
echo "Для остановки нажмите Ctrl+C"


# Quarkus
echo "$(date): Запуск Quarkus конфигурации..."
docker-compose --env-file quarkus.env --env-file .env up --build -d
sleep 11m
echo "$(date): Остановка Quarkus конфигурации..."
docker compose --env-file quarkus.env --env-file .env down

# Micronaut
echo "$(date): Запуск Micronaut конфигурации..."
docker compose --env-file micronaut.env --env-file .env up --build -d
sleep 11m
echo "$(date): Остановка Micronaut конфигурации..."
docker compose --env-file micronaut.env --env-file .env down

# Spring
echo "$(date): Запуск Spring конфигурации..."
docker compose --env-file spring.env --env-file .env up --build -d
sleep 11m
echo "$(date): Остановка Spring конфигурации..."
docker compose --env-file spring.env --env-file .env down
