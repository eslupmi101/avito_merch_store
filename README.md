# Avito Merch Store

## 📌 Описание
**Avito Merch Store** — это сервис, написанный на **Golang** с использованием **паттерна MVC**. Он предназначен для управления продажей мерча.

## 📋 Технологии
- **Golang** 1.22
- **Python** 3.10
- **Docker**

## 🚀 Запуск проекта

Перед запуском создайте файлы окружения:
```sh
cp .env.example .env
```

Затем выполните команду:
```sh
docker compose -f docker-compose.yaml up -d
```

## 🧪 Тестирование
Запустить **юнит-тесты**:
```sh
docker compose -f docker-compose.yaml run --rm app test-unit
```

Запустить **интеграционные тесты**:
```sh
docker compose -f docker-compose.yaml run --rm app test-integration
```

Запустить **все тесты**:
```sh
docker compose -f docker-compose.yaml run --rm app test-all
```

