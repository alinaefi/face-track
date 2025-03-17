# Application Setup

### The application uses several environment variables to manage database connections, server settings, authentication, and integration with external APIs. Below is a detailed description of each environment variable:

### Database Settings
#### These variables are used to configure the PostgreSQL database connection:

FACE_TRACK__PG_HOST    Hostname or IP address of the PostgreSQL server

FACE_TRACK__PG_PORT    Port number on which PostgreSQL is running

FACE_TRACK__PG_USER    Username for connecting to PostgreSQL

FACE_TRACK__PG_PASS    Password for PostgreSQL user

FACE_TRACK__PG_NAME    Name of the database to connect to

### Application Server Settings
#### These variables are used to configure the application server:

FACE_TRACK__SERVER_ADDRESS    Port on which the application server will run.

### Basic Auth Settings for the Application
#### These variables are used to configure the application's basic authentication:

FACE_TRACK__API_USER    Username for Basic Auth

FACE_TRACK__API_PASS    Password for Basic Auth

### External Face Cloud API Settings
#### These variables are used to integrate with the external Face Cloud API:

FACE_CLOUD__API_URL    URL to access the Face Cloud API

FACE_CLOUD__API_USER    Username for authentication in the Face Cloud API (email)

FACE_CLOUD__API_PASS    Password for authentication in the Face Cloud API

**To edit environment variables, refer to the `.env` file**

# Application Startup Guide

## 1. Setting Up Environment Variables

To assign the environment variables required for the application to work, refer to the `.env` file located in the project's root folder.

## 2. Preparing the Container

To create an image, navigate to the project's root directory and run the following command in the terminal:
```
    docker-compose -f deployments/docker-compose.yaml up --build
```

## 3. Starting the Container

To start the container, run the following command in the terminal:
```
    docker-compose -f deployments/docker-compose.yaml up
```
Dependency updates, database connection, database migration, and application server startup should occur automatically.
If you encounter any issues, you can run the application without a container. See section 4.

## 4. Working with the Application

To send requests to the service, you can use tools such as **Postman**, **cURL**, or any other HTTP clients.

## 5. Running the Application Without a Container

If you have trouble preparing or starting the container, you can run the application without it:

### 1. Install Dependencies

    - Go 1.21
    - PostgreSQL
    - Golang migrate CLI (https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### 2. Ensure Environment Variables Are Initialized and Assigned

To assign the required environment variables, refer to the `.env` file in the project's root folder.

### 3. Database Migration

Use the Go migrate CLI command to migrate the PostgreSQL database:
```
    migrate -path=internal/database/migrations -database "postgresql://${FACE_TRACK__PG_USER}:${FACE_TRACK__PG_PASS}@${FACE_TRACK__PG_HOST}:${FACE_TRACK__PG_PORT}/${FACE_TRACK__PG_NAME}?sslmode=disable" up
```
Replace the environment variable names with their corresponding values.

### 4. Start the Application Server

Run the following commands:

To update dependencies:
```
    go mod tidy
```
To compile:
```
    go build
```
To start the application (Linux):
```
    ./face-track
```
Or on Windows:
```
    face-track.exe
```

# API Documentation

### Authentication

Basic authentication is required for access to all endpoints:

- **Username**: ${FACE_TRACK__API_USER}
- **Password**: ${FACE_TRACK__API_PASS}

**To edit environment variables, refer to the `.env` file**

## 1. Adding a Task

### Endpoint

**POST** `/api/tasks`

### Description

Creates a new task and returns the ID of the created task.

### Request

**Method**: `POST`

**Headers**:

- `Content-Type: application/json`
- `Authorization: Basic <base64_encoded_credentials>`

**Example Request**:

```http
POST /api/tasks HTTP/1.1
Host: 127.0.0.1:4221
Authorization: Basic <base64_encoded_credentials>
Content-Type: application/json

{}
```

### Response

#### 1. Error Response

**Code**: `500 Internal Server Error`

**Body**:

- `Content-Type: application/json`

**Example Response**:

```json
{
    "error": "failed to create a task"
}
```

#### 2. Successful Response

**Code**: `200 OK`

**Body**:

- `Content-Type: application/json`

**Example Response**:

```json
{     
    "taskId": 6
}
```

## 2. Getting a Task by ID

### Endpoint

**GET** `/api/tasks/:id`

### Description

Returns task data, images, and associated faces for the specified ID.

### Request

**Method**: `GET`

**Headers**:

- `Content-Type: application/json`
- `Authorization: Basic <base64_encoded_credentials>`

**Example Request**:

```http
GET /api/tasks/{id} HTTP/1.1
Host: 127.0.0.1:4221
Authorization: Basic <base64_encoded_credentials>
{}
```

### Response

#### 1. Error Response

**Code**: `400 Bad Request`

**Body**:

- `Content-Type: application/json`

**Example Response**:

```json
{
    "error": "task with id {id} not found"
}
```

#### 2. Successful Response

**Code**: `200 OK`

**Content-Type**: `application/json; charset=utf-8`

**Response Body**:

```json
{
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
        }
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

## 3. Deleting a Task by ID

### Endpoint

**DELETE** `/api/tasks/:id`

### Description

Deletes task data, images, and associated faces by the specified ID, if the task is not in processing status. Deletes task images from disk.

### Request

**Method**: `DELETE`

**Headers**:

- `Content-Type: application/json`
- `Authorization: Basic <base64_encoded_credentials>`

**Example Request**:

```http
DELETE /api/tasks/{id} HTTP/1.1
Host: 127.0.0.1:4221
Authorization: Basic <base64_encoded_credentials>
{}
```

### Response

#### 1. Error Response

**Code**: `400 Bad Request`

**Example Response**:

```json
{
    "error": "unable to delete task: processing is in progress"
}
```

#### 2. Successful Response

**Code**: `200 OK`

**Example Response**:

```json
{
    "message": "task was successfully deleted"
}
```

## 4. Adding an Image to a Task by ID

### Endpoint

**PATCH** `/api/tasks/:id`

### Description

Adds an image to a task if the task status is "new". Saves the image to disk and stores its data in the database.
Only .JPEG images are allowed.

### Request

**Method**: `PATCH`

**Headers**:

- `Content-Type: multipart/form-data`
- `Authorization: Basic <base64_encoded_credentials>`

**Request Body**:

The form data sent in the request should contain the following fields:

- `image`: image file (only `.jpeg`).
- `imageName`: name of the image (unique for the task; does not necessarily have to match the file name).

### Response

### 1. Error Response

**Code**: `500 Internal Server Error`

**Body**:

- `Content-Type: application/json`

**Example Response**:

```
{
    "error": "failed to add image to task"
}
```

### 2. Successful Response

**Status Code**: `200 OK`

**Content-Type**: `application/json; charset=utf-8`

**Response Body**:

```
{
    "message": "image was successfully added to task"
}
```


## 5. Start Task Processing

### Endpoint

**PATCH** `/api/tasks/:id/process`

### Description

Starts processing task images if the task status is "new" or "error".

### Request

**Method**: `PATCH`

**Headers**:

- `Content-Type: application/json`
- `Authorization: Basic <base64_encoded_credentials>`

**Example Request**:

```http
PATCH /api/tasks/{id}/process HTTP/1.1
Host: 127.0.0.1:4221
Authorization: Basic <base64_encoded_credentials>
{}
```

### Response

### 1. Error Response

**Code**: `400 Bad Request`

**Body**:

- `Content-Type: application/json`

**Example Response**:

```
{
    "error": "invalid task id format"
}
```

### 2. Successful Response

**Status Code**: `200 OK`

**Content-Type**: `application/json; charset=utf-8`

**Response Body**:

```
{
   "message": "task is being processed"
}
```

