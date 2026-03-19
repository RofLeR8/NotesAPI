# NotesAPI

Go REST API для управления заметками с использованием Gin, GORM и PostgreSQL.

## Возможности

- Регистрация и аутентификация пользователей (JWT)
- CRUD операции для заметок
- Защита маршрутов через JWT middleware
- Health check endpoint

## Требования

- Go 1.26+
- Docker и Docker Compose
- PostgreSQL 15+

## Быстрый старт

### 1. Клонирование и настройка

```bash
git clone <repository-url>
cd NotesAPI
cp .env.example .env
```

### 2. Запуск через Docker Compose

```bash
make docker-up
```

Или вручную:

```bash
docker-compose up -d
```

### 3. Локальная разработка

```bash
# Установка зависимостей
go mod download

# Запуск
make run
# или
go run ./cmd/api
```

## API Endpoints

### Health Check

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/health` | Проверка статуса API |

**Ответ:**
```json
{ "status": "ok" }
```

---

### Аутентификация

#### Регистрация

| Метод | Endpoint |
|-------|----------|
| POST | `/api/users/register` |

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Ответ (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

#### Login

| Метод | Endpoint |
|-------|----------|
| POST | `/api/users/login` |

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Ответ (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

---

### Заметки (требуется аутентификация)

Все запросы к заметкам требуют заголовок:
```
Authorization: Bearer <token>
```

#### Получить все заметки

| Метод | Endpoint |
|-------|----------|
| GET | `/api/notes` |

**Ответ:**
```json
[
  {
    "id": 1,
    "title": "Моя заметка",
    "content": "Содержимое заметки",
    "user_id": 1,
    "created_at": "2025-01-01T12:00:00Z",
    "updated_at": "2025-01-01T12:00:00Z"
  }
]
```

#### Получить одну заметку

| Метод | Endpoint |
|-------|----------|
| GET | `/api/notes/:id` |

**Ответ:**
```json
{
  "id": 1,
  "title": "Моя заметка",
  "content": "Содержимое заметки",
  "user_id": 1,
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:00Z"
}
```

#### Создать заметку

| Метод | Endpoint |
|-------|----------|
| POST | `/api/notes` |

**Request:**
```json
{
  "title": "Новая заметка",
  "content": "Содержимое новой заметки"
}
```

**Ответ (201):**
```json
{
  "id": 2,
  "title": "Новая заметка",
  "content": "Содержимое новой заметки",
  "user_id": 1,
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:00Z"
}
```

#### Обновить заметку

| Метод | Endpoint |
|-------|----------|
| PUT | `/api/notes/:id` |

**Request:**
```json
{
  "title": "Обновленный заголовок",
  "content": "Обновленное содержимое"
}
```

**Ответ (200):**
```json
{
  "id": 1,
  "title": "Обновленный заголовок",
  "content": "Обновленное содержимое",
  "user_id": 1,
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:00Z"
}
```

#### Удалить заметку

| Метод | Endpoint |
|-------|----------|
| DELETE | `/api/notes/:id` |

**Ответ (200):**
```json
{ "message": "note deleted" }
```

---

### Пользователи (требуется аутентификация)

#### Получить текущего пользователя

| Метод | Endpoint |
|-------|----------|
| GET | `/api/users/:id` |

**Ответ:**
```json
{
  "id": 1,
  "email": "user@example.com"
}
```

---

## Конфигурация

Переменные окружения (см. `.env.example`):

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DB_HOST` | Хост PostgreSQL | localhost |
| `DB_PORT` | Порт PostgreSQL | 5432 |
| `DB_USER` | Пользователь БД | notes_user |
| `DB_PASSWORD` | Пароль БД | notes_password |
| `DB_NAME` | Имя БД | notes_db |
| `PORT` | Порт API | 8080 |
| `JWT_SECRET` | Секретный ключ JWT | change-me-in-production |
| `JWT_EXP_HOURS` | Время жизни токена (часы) | 12 |

## Тесты

```bash
make test
# или
go test -v ./...
```

## Make команды

| Команда | Описание |
|---------|----------|
| `make build` | Собрать бинарный файл |
| `make run` | Запустить приложение |
| `make test` | Запустить тесты |
| `make clean` | Очистить артефакты сборки |
| `make docker-build` | Собрать Docker образ |
| `make docker-up` | Запустить Docker Compose |
| `make docker-down` | Остановить Docker Compose |
| `make docker-restart` | Перезапустить Docker Compose |
| `make docker-logs` | Просмотр логов |
| `make docker-clean` | Очистить контейнеры и volumes |

## Структура проекта

```
NotesAPI/
├── cmd/
│   └── api/
│       └── main.go          # Точка входа
├── internal/
│   ├── config/              # Конфигурация
│   ├── handlers/            # HTTP обработчики
│   ├── middleware/          # Gin middleware
│   ├── models/              # GORM модели
│   ├── repository/          # Доступ к БД
│   ├── service/             # Бизнес-логика
│   └── utils/               # Утилиты (JWT, password hashing)
├── migrations/              # SQL миграции
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## License

MIT
