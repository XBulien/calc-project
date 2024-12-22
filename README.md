# Калькулятор

Привет! Это мой проект по созданию микросервиса, способного вычислять значение введенного выражения через POST запрос.

## Структура проекта
calc-project/<br>
├── cmd/<br>
│ └── calc_service/<br>
│     └── main.go<br>
├── internal/<br>
│ └── calculator/<br>
│     └── calculator.go<br>
├── go.mod<br>
└── README.md<br>


- `cmd/calc_service/main.go`: Точка входа в приложение. Содержит логику веб-сервера.
- `internal/calculator/calculator.go`: Пакет с логикой вычисления арифметических выражений.
- `go.mod`: Файл для управления зависимостями Go.
- `README.md`: Документация проекта.

## Установка и запуск

### Требования

- Go 1.23.0 или выше.

### Шаги для запуска

1. Клонируйте репозиторий:

   ```bash
   git clone https://github.com/xbulien/calc-project.git
   cd calc-project
   ```

2. Убедитесь, что файл go.mod содержит правильный путь к модулю:

    ```go
    module github.com/xbulien/calc-project

    go 1.23.0
    ```

3. Запустите сервер:

    ```bash
    go run ./cmd/calc_service/main.go

4. Сервер будет доступен по адресу http://localhost:8080

### Примеры использования

## Успешный запрос

Отправьте POST-запрос с корректным выражением:

```bash
   curl --location 'http://localhost:8080/api/v1/calculate' \
   --header 'Content-Type: application/json' \
   --data '{
    "expression": "2+2*2"
   }'
   ```
Ожидаемый ответ:

```json
   {
      "result": "6.000000"
   }
   ```


## Ошибка 422 (некорректное выражение)

Отправьте POST-запрос с некорректным выражением:

```bash
   curl --location 'http://localhost:8080/api/v1/calculate' \
   --header 'Content-Type: application/json' \
   --data '{
    "expression": "2+a"
   }'
   ```

Ожидаемый ответ:

```json
   {
      "error": "Expression is not valid"
   }
   ```
