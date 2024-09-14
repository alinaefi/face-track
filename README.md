# Настройка приложения

### Приложение использует несколько переменных окружения для управления подключением к базе данных, настройками сервера, аутентификацией и интеграцией с внешним API. Ниже приведено подробное описание каждой переменной окружения:

### Настройки базы данных
#### Эти переменные используются для настройки подключения к базе данных PostgreSQL:

FACE_TRACK__PG_HOST    Имя хоста или IP-адрес сервера PostgreSQL

FACE_TRACK__PG_PORT    Номер порта, на котором работает PostgreSQL

FACE_TRACK__PG_USER    Имя пользователя для подключения к PostgreSQL
 
FACE_TRACK__PG_PASS    Пароль пользователя PostgreSQL

FACE_TRACK__PG_NAME    Имя базы данных, к которой будет выполнено подключение

### Настройки сервера приложения
#### Эти переменные используются для настройки сервера приложения:

FACE_TRACK__SERVER_ADDRESS    Порт, на котором будет запущен сервер приложения.

### Настройки Basic Auth для приложения
#### Эти переменные используются для настройки базовой аутентификации приложения:

FACE_TRACK__API_USER    Имя пользователя для Basic Auth

FACE_TRACK__API_PASS    Пароль для Basic Auth 

### Настройки для работы с внешним API Face Cloud
#### Эти переменные используются для интеграции с внешним API Face Cloud:

FACE_CLOUD__API_URL    URL для доступа к API Face Cloud

FACE_CLOUD__API_USER    Имя пользователя для аутентификации в API Face Cloud (email)

FACE_CLOUD__API_PASS    Пароль для аутентификации в API Face Cloud

**Для редактирования переменных окружения обратитесь к файлу `.env`**


# Инструкция по запуску приложения

## 1. Установка переменных окружения

Для назначения переменных окружения, необходимых для работы приложения, обратитесь к файлу `.env`, который находится в корневой папке проекта.

## 2. Подготовка контейнера

Для создания образа перейдите в корневую директорию проекта и выполните команду в командной строке:
```
    docker-compose build
```

## 2. Запуск образа

Для запуска контейнера выполните команду в командной строке:
```
    docker-compose up
```
Обновление зависимостей, подключение к бд, миграция бд и запуск сервера приложения должны выполняться автоматически.
Если возникли трудности, Вы можете запустить приложение без контейнера, см. пункт 4

## 3. Работа с приложением

Для отправления запросов к сервису вы можете воспользоваться инструментами, такими как **Postman**, **cURL**, или любыми другими HTTP-клиентами.

## 4. Запуск приложения без контейнера

Если у Вас возникли трудности с подготовкой или запуском контейнера, можно запустить приложение без него:

### 1. Установите зависимости

    - Go 1.21
    - PostgreSQL
    - Golang migrate CLI (https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### 2. Убедитесь, что переменные окружения инициализированы и назначены

Для назначения переменных окружения, необходимых для работы приложения, обратитесь к файлу `.env`, который находится в корневой папке проекта.

### 2. Миграция базы данных

Воспользуйтесь командой go migrate CLI для миграции базы данных postrges:
```
    migrate -path=internal/database/migrations -database "postgresql://${FACE_TRACK__PG_USER}:${FACE_TRACK__PG_PASS}@${FACE_TRACK__PG_HOST}:${FACE_TRACK__PG_PORT}/${FACE_TRACK__PG_NAME}?sslmode=disable" up
```
где названия переменных окружения замените на соответствующие значения.

### 3. Запуск сервера приложения

Запустите несколько команд:

Для обновления зависимостей:
```
    go mod tidy
```
Для компиляции:
```
    go build
```
Для запуска (linux)
```
    ./face-track 
```

Или Windows:
```
    face-track.exe
```

# Документация API

### Аутентификация

Для доступа ко всем эндпоинтам требуется базовая аутентификация:

- **Имя пользователя**: ${FACE_TRACK__API_USER}
- **Пароль**: ${FACE_TRACK__API_PASS}

**Для редактирования переменных окружения обратитест к файлу `.env**

## 1. Добавление задания

### Эндпоинт

**POST** `/api/tasks`

### Описание

Создаёт новое задание и возвращает ID созданного задания.

### Запрос

**Метод**: `POST`

**Заголовки**:

- `Content-Type: application/json`
- `Authorization: Basic <base64_encoded_credentials>`

**Пример запроса**:

```http
POST /api/tasks HTTP/1.1
Host: 127.0.0.1:4221
Authorization: Basic <base64_encoded_credentials>
Content-Type: application/json

{}
```

### Ответ

### 1. Ответ с ошибкой

**Код**: `500 Internal Server Error`

**Тело**:

- `Content-Type: application/json`

**Пример ответа**:

```{
    "error": "failed to create a task"
}
```

### 2. Успешный ответ

**Код**: `200 OK`

**Тело**:

- `Content-Type: application/json`

**Пример ответа**:

```
{     
    "taskId": 6

}
```

## 2. Получение задания по ID

### Эндпоинт

**GET** `/api/tasks/:id`

### Описание

Возвращает данные о задании, изображениях и лицах, связанных с ним, по указанному ID.

### Запрос

**Метод**: `GET`

**Заголовки**:

- `Content-Type: application/json`
- `Authorization: Basic <base64_encoded_credentials>`

**Пример запроса**:

```http
GET /api/tasks/{id} HTTP/1.1
Host: 127.0.0.1:4221
Authorization: Basic <base64_encoded_credentials>
{}
```

### Ответ

### 1. Ответ с ошибкой

**Код**: `400 Bad Request`

**Тело**:

- `Content-Type: application/json`

**Пример ответа**:

```
{
    "error": "task with id {id} not found"
}
```

### 2. Успешный ответ

**Код состояния**: `200 OK`

**Content-Type**: `application/json; charset=utf-8`

**Тело ответа**:

```{
    "id": 5,
    "images": [
        {
            "faces": [
                {
                    "age": 28,
                    "bbox": {
                        "height": 380,
                        "width": 269,
                        "x": 175,
                        "y": 66
                    },
                    "gender": "male"
                }
            ],
            "name": "collage1"
        },
    ],
    "statistics": {
        "ageFemaleAvg": 0,
        "ageMaleAvg": 28,
        "facesFemale": 0,
        "facesMale": 1,
        "facesTotal": 1
    },
    "taskStatus": "completed"
}
```

## 3. Удаление задания по ID

### Эндпоинт

**DELETE** `/api/tasks/:id`

### Описание

Удаляет данные о задании, изображениях и лицах, связанных с ним, по указанному ID, если задание не в статусе обработки.
Удаляет изображения задания на диске.

### Запрос

**Метод**: `DELETE`

**Заголовки**:

- `Content-Type: application/json`
- `Authorization: Basic <base64_encoded_credentials>`

**Пример запроса**:

```http
DELETE /api/tasks/{id} HTTP/1.1
Host: 127.0.0.1:4221
Authorization: Basic <base64_encoded_credentials>
{}
```

### Ответ

### 1. Ответ с ошибкой

**Код**: `400 Bad Request`

**Тело**:

- `Content-Type: application/json`

**Пример ответа**:

```
{
    "error": "unable to delete task: processing is in progress"
}
```

### 2. Успешный ответ

**Код состояния**: `200 OK`

**Content-Type**: `application/json; charset=utf-8`

**Тело ответа**:

```
{
    "message": "task was successfully deleted"
}
```


## 4. Добавление изображения к заданию по ID

### Эндпоинт

**PATCH** `/api/tasks/:id`

### Описание

Добавляет изображение к заданию, если статус задания "new". Сохраняет изображение на диске и данные о нем в базе данных.
Допускаются только .JPEG изображения.

### Запрос

**Метод**: `PATCH`

**Заголовки**:

- `Content-Type: multipart/form-data`
- `Authorization: Basic <base64_encoded_credentials>`


**Тело запроса**:

Форма данных, отправляемая в запросе, должна содержать следующие поля:

- `image`: файл изображения (только `.jpeg`).
- `imageName`: название изображения (уникальное для задания; не обязательно совпадающее с именем файла).

### Ответ

### 1. Ответ с ошибкой

**Код**: `500 Internal Server Error`

**Тело**:

- `Content-Type: application/json`

**Пример ответа**:

```
{
    "error": "failed to add image to task"
}
```

### 2. Успешный ответ

**Код состояния**: `200 OK`

**Content-Type**: `application/json; charset=utf-8`

**Тело ответа**:

```
{
    "message": "image was successfully added to task"
}
```


## 5. Запуск обработки задания

### Эндпоинт

**PATCH** `/api/tasks/:id/process`

### Описание

Запускает обработку изображений задания, если статус задания "new" или "error". 

### Запрос

**Метод**: `PATCH`

**Заголовки**:

- `Content-Type: application/json`
- `Authorization: Basic <base64_encoded_credentials>`


**Пример запроса**:

```http
PATCH /api/tasks/{id}/process HTTP/1.1
Host: 127.0.0.1:4221
Authorization: Basic <base64_encoded_credentials>
{}
```

### Ответ

### 1. Ответ с ошибкой

**Код**: `400 Bad Request`

**Тело**:

- `Content-Type: application/json`

**Пример ответа**:

```
{
    "error": "invalid task id format"
}
```

### 2. Успешный ответ

**Код состояния**: `200 OK`

**Content-Type**: `application/json; charset=utf-8`

**Тело ответа**:

```
{
   "message": "task is being processed"
}
```




