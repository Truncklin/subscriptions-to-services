# Subscriptions to Services

REST-сервис для управления информацией о подписках пользователей:  
CRUDL-операции (Create, Read, Update, Delete, List) + фильтрация по периоду.  
Сервис написан на Go, использует PostgreSQL, миграции, логирование и запускается с помощью Docker Compose.

---

## 🚀 Возможности

- Создание подписки с датой начала/окончания  
- Получение подписки по ID  
- Обновление подписки  
- Удаление подписки  
- Список подписок с фильтрацией по периоду, `user_id` и названию сервиса  
- Подсчёт суммарной стоимости подписок за период  
- Хранение данных в PostgreSQL

---

## 📦 Технологии

- Go (Golang)
- PostgreSQL
- pgxpool
- chi (во внутреннем роутере)
- Docker / Docker Compose
- Swagger (OpenAPI)
- Миграции SQL

---

## 📥 Установка

Склонируйте репозиторий:

```bash
git clone https://github.com/Truncklin/subscriptions-to-services.git
cd subscriptions-to-services
```

## 🔧 Настройка

Создайте файл конфигурации:

`configs/local.yaml`:

```yaml
env: "localhost"
storage_path: "postgres://prsvc:prsvcpass123@db:5432/db?sslmode=disable"
http_server:
  host: "0.0.0.0:8080"
  timeout: 10s
  idle_timeout: 60s
```

## 🐳 Запуск через Docker Compose

Убедитесь, что установлены Docker и Docker Compose.

Соберите и запустите сервис:

```bash
docker compose up --build
```

Это поднимет:

- PostgreSQL (контейнер `db`)
- Ваш сервис (контейнер `subservices_app`)

📌 Примеры API (Postman / curl)

Создать подписку

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025",
    "end_date": "12-2025"
  }'
```

Получить подписку

```bash
curl http://localhost:8080/api/v1/subscriptions/{id}
```

Обновить подписку

```bash
curl -X PUT http://localhost:8080/api/v1/subscriptions/{id} \
  -H "Content-Type: application/json" \
  -d '{
      "service_name": "Yandex Plus",
      "price": 450,
      "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
      "start_date": "07-2025",
      "end_date": "01-2026"
  }'
```

Удалить подписку

```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions/{id}
```

Список подписок с фильтрацией

```bash
curl "http://localhost:8080/api/v1/subscriptions?from=01-2025&to=12-2025"
```

##📊 Swagger / OpenAPI

Если в проекте настроен Swagger через swag и подключён в сервере, открыть документацию можно по URL:

http://localhost:8080/swagger/index.html

##🗃 Структура проекта

```
├── cmd/webserver          — точка входа
├── internal/config        — загрузка конфигурации из YAML
├── internal/http/handlers — HTTP‑ручки
├── internal/http/router   — маршруты
├── internal/storage       — миграции и подключение к БД
├── configs                — конфиги для запуска
├── Dockerfile
├── docker-compose.yaml
└── README.md
```

🛠 Миграции

Миграции находятся в папке:
```
internal/storage/migrations/
```
Они создают таблицу subscriptions и нужные индексы.
