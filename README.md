# test-config-server
Сервис-конфигуратор, тестовая задача для Golang разработчика. Текст задания в файле challenge_go.txt.

## Сборка, настройка, запуск и тесты
1. Склонировать репозитарий
2. `go build`
3. Задать [строку параметров соединения с базой данных](https://godoc.org/github.com/lib/pq) через переменную окружения **TEST_CONFIG_DB** 
4. Запустить миграции `test-config-server migrate`
    - Опционально запустить тесты `go test . ./migration`. Тесты используют данные, занесенные в базу на этапе миграции(см. замечания)
5. Запустить сам сервис `test-config-server run`. По умолчнию сервис слушает на ':8081', можно настроить через переменную **TEST_CONFIG_ADDR**

## Пример запроса и ответа
POST запрос в корень http-сервера: `{"Type": "database.postgres", "Data": "service.test"}`

Ответ(фактчески ответ отдаётся без переносов строк и отступов):
``
{
    "host": "localhost",
    "port": "5432",
    "database": "devdb",
    "user": "mr_robot",
    "password": "secret",
    "schema": "public"
}
``

- Если тело запроса не является валидным JSON или поля *Type*/*Data* отсутствуют или заданы пустыми строками, возвращается ошибка 400.
- Если данные не найдены в базе, возвращается 404.
- В случае проблем с базой данных, может вовращаться 500.

## Замечания
- Возможно, задание предполагало создание отдельных таблиц для каждого типа конифгурации ради снижения вероятности ошибок и упрощения параметрического редактирования(массовая смена хоста при переезде базы данных, например).
    * Думаю, что данные стоит валидировать до давления в базу.
    * Свежий Postgres предоставляет достаточно возможностей для работы с JSON.
- Судя по всему, что предполагалось использование **внешнего** модуля миграций... Внутренний вариант был выбран в основном по привычке и невнимательности, но у него есть и плюсы(в общем случае): так можно легко использовать части внутренней логики сервиса.