# Backend для аналога Instagram

Backend-приложение на Go, реализующее функциональность социальной сети, аналогичной Instagram. Проект предоставляет REST API для управления пользователями, публикациями, комментариями, лайками, подписками и уведомлениями в реальном времени через WebSocket.

## Команда

- [Белей Егор](https://vk.com/snprmcs)
- [Шевцов Данил](https://vk.com/l_rave_l)
- [Субботин Кирилл](https://vk.com/id323489516)

## Ссылки

- [Frontend репозиторий](https://github.com/DeniedSaBUro/MPT_Project_frontend)

## Содержание

- [Обзор](#обзор)
- [Технологии](#технологии)
- [Файловая структура проекта](#структура-проекта)
- [Установка и настройка](#установка-и-настройка)
- [Запуск](#запуск)
- [API эндпоинты](#api-эндпоинты)
- [Тестирование](#тестирование)
- [Разработка](#разработка)
- [Архитектура](#архитектура)

## Обзор

Проект представляет собой REST API сервер, который обеспечивает:

1. Регистрация и авторизация пользователей
2. Создание, редактирование, удаление постов с медиа
3. Подписки на пользователей
4. Лайки постов
5. Комментарии к постам и ответы на них
6. Система уведомлений о лайках, постах и комментариях

## Технологии

- **Go 1.25.3** - основной язык программирования
- **PostgreSQL** - база данных
- **JWT** (golang-jwt/jwt/v5) - аутентификация
- **WebSocket** (gorilla/websocket) - уведомления в реальном времени
- **UUID** (google/uuid) - генерация уникальных идентификаторов
- **bcrypt** (golang.org/x/crypto) - хеширование паролей
- **godotenv** - управление переменными окружения

## Файловая структура проекта
> **Примечание**: для content, comment, like, follow, notification большая часть структуры пропущена, так как она построена по аналогии с user

```
project/
├── cmd/
│   └── app/
│       └── main.go              # Точка входа приложения
│
├── docs/                        # Сгенерированные файлы сваггера
│   ├── docs.go                  # Настройки сваггера
│   ├── swagger.json             # Спецификация (JSON)
│   └── swagger.yaml             # Спецификация (YAML)
│
├── internal/                    # Внутренние пакеты
│   ├── auth/                    # Аутентификация и авторизация
│   │   ├── authHandler.go       # HTTP обработчик
│   │   ├── authService.go       # Бизнес-логика
│   │   ├── jwt.go               # Работа с JWT токенами
│   │   ├── loginRequest.go      # Объект для передачи данных для логина пользователя
│   │   ├── loginResponse.go     # Объект для возврата токена
│   │   └── middleware.go        # Middleware для проверки токенов
│   │
│   ├── user/                    # Управление пользователями
│   │   ├── userHandler.go                # HTTP обработчик
│   │   ├── userService.go                # Бизнес-логика
│   │   ├── userServiceInterface.go       # Методы бизнес-логики
│   │   ├── userRepository.go             # Работа с БД
│   │   ├── userRepositoryInterface.go    # Методы работы с БД
│   │   ├── userModel.go                  # Модель данных
│   │   ├── userDTO.go                    # Объект для передачи данных о регистрации пользователя
│   │   ├── registerDTO.go                # Объект для отображения данных о пользователе
│   │   └── updatePasswordDTO             # Объект для передачи данных о старом и новом пароле
│   │
│   ├── content/                 # Управление постами
│   ├── comment/                 # Комментарии к постам
│   ├── like/                    # Лайки постов
│   ├── follow/                  # Подписки на пользователей
│   ├── notification/            # Уведомления
│   │   └── webSocketHub.go      # Регистрация соединений и отправка сообщений через webSocket
│   │
│   ├── storage/                          # Работа с хранением файлов
│   │   ├── avatarStorage.go              # Хранение аватаров
│   │   ├── avatarStorageInterface.go     # Методы хранения аватаров
│   │   ├── mediaStorageInterface.go      # Методы хранения медиа-файлов для постов
│   │   └── mediaStorage.go               # Хранение медиа-файлов для постов
│   │
│   ├── config/                  # Конфигурация
│   │   └── config.go            # Загрузка конфигурации с использованием .env
│   │   
│   ├── database/                # Работа с БД
│   │   ├── database.go          # Подключение базы данных
│   │   └── schema.sql           # SQL схема БД
│   │
│   ├── cors/                    # CORS настройки
│   │   ├── CORShandler.go       # middleware для CORS
│   │   └── enableCORS.sql       # Настройки конфигурации CORS
│   │
│   └── contextKeys/contextKeys.go        # Контекстные ключи
│
├── uploads/                     # Загруженные файлы
│   ├── avatars/                 # Аватары пользователей
│   └── content/                 # Медиа постов
│
├── .env                         # Переменные среды
├── .env.example                 # Пример переменных среды
├── go.mod                       # Зависимости Go
├── go.sum                       # Контрольные суммы зависимостей
└── README.md                    # Документация
```

## Установка и настройка

1. **Клонируйте репозиторий** (если еще не склонирован):
   ```bash
   git clone <repository-url>
   cd <project>
   ```

2. **Установите зависимости**:
   ```bash
   go mod download
   ```

3. **Настройте базу данных PostgreSQL**:
   - Создайте базу данных:
     ```sql
     CREATE DATABASE zapretgram_db;
     ```
   - Примените схему из `internal/database/schema.sql`

4. **Создайте файл `.env`** в корне проекта:
   ```env
   DB_URL=postgres://user:password@server:port/db_name?sslmode=disable
   JWT_SECRET=your-secret-key-here-min-32-chars
   TOKEN_TTL=24h
   ```
   
   Где:
   - `DB_URL` - строка подключения к PostgreSQL
   - `JWT_SECRET` - секретный ключ для подписи JWT токенов
   - `TOKEN_TTL` - время жизни токена (например, `24h`, `168h` для недели)

   Более подробная информация есть в `.env.example`

## Запуск

1. Убедитесь, что PostgreSQL запущен и база данных создана
2. Проверьте, что файл `.env` настроен корректно
3. Запустите приложение:
   ```bash
   go run cmd/app/main.go
   ```

4. Сервер запустится на `http://localhost:8080`

## API эндпоинты

Вся информация о эндпоинтах доступна после запуска сервера по адресу `http://localhost:8080/swagger`

## Тестирование

Проект содержит комплексное тестовое покрытие основных компонентов системы. Все тесты написаны с использованием стандартной библиотеки `testing` Go и дополнительных библиотек для моков и assertions.

### Используемые библиотеки для тестирования

- **testing** - стандартная библиотека Go для тестирования
- **stretchr/testify/assert** - библиотека для удобных assertions
- **stretchr/testify/mock** - библиотека для создания моков
- **DATA-DOG/go-sqlmock** - библиотека для мокирования SQL запросов
- **net/http/httptest** - стандартная библиотека для HTTP тестирования

### Структура тестов

Тесты организованы по модулям и следуют принципу "один файл тестов на один файл исходного кода". Файлы с тестами имеют суффикс `_test.go` и находятся в тех же пакетах, что и тестируемый код.

```
internal/
├── user/
│   ├── userService_test.go       # Тесты для бизнес-логики пользователей
│   ├── userRepository_test.go    # Тесты для работы с БД (моки SQL)
│   └── userHandler_test.go       # Тесты для HTTP handlers
├── auth/
│   ├── authService_test.go       # Тесты для логики аутентификации
│   ├── authHandler_test.go       # Тесты для HTTP handlers авторизации
│   ├── jwt_test.go               # Тесты для JWT токенов
│   └── middleware_test.go        # Тесты для middleware
├── content/
│   ├── contentService_test.go
│   ├── contentRepository_test.go
│   └── contentHandler_test.go
├── storage/
│   ├── avatarStorage_test.go     # Тесты для работы с файлами аватаров
│   └── mediaStorage_test.go      # Тесты для работы с медиа файлами
└── ...
```

### Типы тестов

#### 1. Тесты Services (Бизнес-логика)

Тесты для сервисов проверяют бизнес-логику приложения. Используются моки репозиториев для изоляции от базы данных.

**Пример структуры:**
```go
// Создание мока репозитория
type TestUserRepository struct {
    usersByID map[uuid.UUID]*User
    // ...
}

// Тестирование сценариев
func TestUserService_Create_EdgeCases(t *testing.T) {
    t.Run("Ошибка при пустом пароле", func(t *testing.T) {
        // Arrange
        repo := NewTestUserRepository()
        service := createTestUserService(repo)
        
        // Act
        err := service.Create(context.Background(), user)
        
        // Assert
        if err == nil {
            t.Error("Create() должен вернуть ошибку")
        }
    })
}
```

**Что тестируется:**
- Валидация входных данных
- Обработка ошибок репозитория
- Бизнес-правила и логика
- Edge cases (граничные случаи)

#### 2. Тесты Repositories (Работа с БД)

Тесты для репозиториев проверяют корректность SQL запросов. Используется `go-sqlmock` для мокирования базы данных.

**Пример структуры:**
```go
func TestUserRepository_Create(t *testing.T) {
    // Создание мока БД
    db, mock, err := sqlmock.New()
    defer db.Close()
    
    repo := NewUserRepository(db)
    
    t.Run("Успешное создание пользователя", func(t *testing.T) {
        // Настройка ожиданий SQL запроса
        mock.ExpectExec("INSERT INTO users").
            WithArgs(userID, "testuser", "hashedpassword", createdAt).
            WillReturnResult(sqlmock.NewResult(1, 1))
        
        // Выполнение операции
        err := repo.Create(context.Background(), user)
        
        // Проверка выполнения ожиданий
        if err := mock.ExpectationsWereMet(); err != nil {
            t.Errorf("невыполненные ожидания: %v", err)
        }
    })
}
```

**Что тестируется:**
- Корректность SQL запросов
- Передача параметров в запросы
- Обработка ошибок БД
- Правильная обработка результатов запросов

#### 3. Тесты Handlers (HTTP endpoints)

Тесты для handlers проверяют обработку HTTP запросов. Используется `httptest` для создания тестовых запросов и записи ответов.

**Пример структуры:**
```go
func TestAuthHandler_Login_Success(t *testing.T) {
    // Создание мока сервиса
    testService := new(TestAuthService)
    handler := NewAuthHandler(testService)
    
    // Настройка ожиданий
    testService.On("Login", mock.Anything, "testuser", "password123").
        Return("test.jwt.token", nil)
    
    // Создание тестового запроса
    reqBody := map[string]string{
        "username": "testuser",
        "password": "password123",
    }
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    // Выполнение handler
    handler.Login(w, req)
    
    // Проверка результата
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
    
    // Проверка вызова моков
    testService.AssertExpectations(t)
}
```

**Что тестируется:**
- HTTP статус коды ответов
- Формат JSON ответов
- Валидация входных данных
- Обработка различных HTTP методов
- Корректность вызова сервисов

#### 4. Тесты Storage (Работа с файлами)

Тесты для storage проверяют сохранение и чтение файлов. Используются временные директории для изоляции тестов.

**Пример структуры:**
```go
func TestFileAvatarStorage_Save(t *testing.T) {
    // Создание временной директории
    tmpDir, err := os.MkdirTemp("", "avatar-storage-test-*")
    defer os.RemoveAll(tmpDir)
    
    storage := NewAvatarStorage(tmpDir, "/uploads")
    
    // Тестирование сохранения
    imageData := "fake image data"
    reader := strings.NewReader(imageData)
    
    url, err := storage.Save(context.Background(), userID, reader)
    // Проверка результата
}
```

**Что тестируется:**
- Сохранение файлов
- Генерация путей к файлам
- Обработка ошибок при работе с файловой системой

### Запуск тестов

#### Запуск всех тестов:
```bash
go test ./...
```

#### Запуск тестов конкретного пакета:
```bash
go test ./internal/user
go test ./internal/auth
```

#### Запуск конкретного теста:
```bash
go test -run TestUserService_Create ./internal/user
```

#### Запуск с подробным выводом:
```bash
go test -v ./...
```

#### Запуск с покрытием кода:
```bash
go test -cover ./...
```

#### Генерация отчета о покрытии:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Запуск тестов с таймаутом:
```bash
go test -timeout 30s ./...
```

### Стратегия тестирования

1. **Unit тесты** - тестируют отдельные функции и методы в изоляции
2. **Интеграционные тесты** - тестируют взаимодействие между компонентами (репозитории с БД через моки)
3. **Handler тесты** - тестируют HTTP эндпоинты в изоляции от реальных сервисов

### Покрытие кода

Проект стремится к высокому покрытию тестами всех критических компонентов:
- Services - проверка бизнес-логики
- Repositories - проверка SQL запросов
- Handlers - проверка HTTP обработки
- Storage - проверка работы с файлами
- JWT и middleware - проверка аутентификации и авторизации

### Примеры тестовых сценариев

#### Тестирование успешных сценариев
- Создание пользователя с валидными данными
- Успешный вход в систему
- Получение данных пользователя
- Обновление профиля

#### Тестирование ошибок
- Валидация входных данных (пустые поля, неверный формат)
- Ошибки базы данных
- Ошибки файловой системы
- Неверные учетные данные
- Отсутствие ресурсов (пользователь не найден)

#### Тестирование граничных случаев
- Пустые строки и nil значения
- Максимальная длина полей
- Специальные символы в строках
- Дублирование данных (например, username)

### Эффективность

1. **Изоляция тестов** - каждый тест должен быть независимым
2. **Моки для зависимостей** - использование моков для изоляции от внешних систем
3. **Именование тестов** - понятные имена, описывающие тестируемый сценарий
4. **Table-driven tests** - использование табличных тестов для множественных сценариев
5. **Очистка ресурсов** - использование defer для очистки временных ресурсов
6. **Проверка всех аспектов** - не только успешные сценарии, но и обработка ошибок

## Разработка

### Соглашения о коде

1. **Именование**:
   - Handlers: `*Handler`
   - Services: `*Service`
   - Repositories: `*Repository`
   - Models: `*Model`
   - DTOs: `*DTO`/`*Response`/`*Request`

2. **Структура файлов в модуле**:
   - `*Handler.go` - HTTP обработчики
   - `*Service.go` - бизнес-логика
   - `*ServiceInterface.go` - интерфейс сервиса
   - `*Repository.go` - работа с БД
   - `*RepositoryInterface.go` - интерфейс репозитория
   - `*Model.go` - модели данных
   - `*DTO.go` - объекты для передачи или отображения данных

3. **Обработка ошибок**:
   - Использование стандартных HTTP статусов
   - Возврат понятных сообщений об ошибках

### Добавление нового функционала

1. **Создать новый модуль**:
   ```
   internal/newfeature/
   ├── newfeatureHandler.go
   ├── newfeatureService.go
   ├── newfeatureServiceInterface.go
   ├── newfeatureRepository.go
   ├── newfeatureRepositoryInterface.go
   └── newfeatureModel.go
   ```

2. **Зарегистрировать в main.go**:
   - Создать репозиторий, сервис, handler
   - Добавить маршруты в `mux`

3. **Обновить схему БД** (если нужно):
   - Добавить схему в `internal/database/schema.sql`

4. **Добавить новый middleware** (если нужно):
   - Последовательно обернуть `corsMux` в main.go новыми middleware

### Работа с базой данных

- **Схема**: SQL файл в `internal/database/schema.sql`
- **Подключение**: Настройка через `DB_URL` в `.env`
- **Пул соединений**: Настроен в `main.go` (MaxOpenConns: 10, MaxIdleConns: 5)

### Работа с файлами

- **Аватары**: Сохраняются в `uploads/avatars/`
- **Медиа постов**: Сохраняются в `uploads/content/{postId}/`
- **Имена файлов**: Используются UUID для уникальности
- **Изменение варианта хранения**: Использование абстракции позволяет реализовать другой вид хранения, например S3

### WebSocket уведомления

- **Хаб**: Централизованный хаб для управления соединениями
- **Аутентификация**: Токен передается через заголовок Authenticate
- **Типы уведомлений**: Лайки, комментарии, подписки

## Архитектура

Проект следует принципам чистой архитектуры с разделением на слои:

```
┌─────────────────────────────────────┐
│         HTTP Handlers               │  ← Слой представления
├─────────────────────────────────────┤
│         Services                    │  ← Бизнес-логика
├─────────────────────────────────────┤
│         Repositories                │  ← Доступ к данным
├─────────────────────────────────────┤
│         Database                    │  ← Хранилище
└─────────────────────────────────────┘
```

### Основные компоненты

1. **cmd/app/main.go** - Точка входа:
   - Инициализация конфигурации
   - Подключение к БД
   - Создание зависимостей (repositories, services, handlers)
   - Регистрация маршрутов
   - Регистрация middleware
   - Запуск HTTP сервера

2. **internal/config** - Конфигурация:
   - Загрузка переменных окружения
   - Настройки БД, JWT

3. **internal/auth** - Аутентификация:
   - JWT генерация и валидация
   - Middleware для защиты маршрутов
   - Логин

4. **internal/user** - Пользователи:
   - Управление профилями
   - Загрузка аватаров

5. **internal/content** - Посты:
   - Создание и управление постами
   - Загрузка медиа-файлов

6. **internal/notification** - Уведомления:
   - Хаб WebSocket
   - Отправка уведомлений в реальном времени
   - Хранение уведомлений в БД

7. **internal/storage** - Хранилище:
   - Абстракция для работы с файлами
   - Сохранение аватаров и медиа

### Потоки данных

**Создание поста**:
```
HTTP Request → Handler → Service → Repository → Database
                            ↓
                        Storage (медиа)
```

**WebSocket уведомления**:
```
Event → Service → NotificationService → WebSocketHub → Clients
```

### Безопасность

- **JWT токены**: Используются для аутентификации
- **Хеширование паролей**: bcrypt для безопасного хранения
- **CORS**: Настроен для кросс-доменных запросов

**Разработано для учебного проекта по дисциплине "Современные технологии программирования"**
