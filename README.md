# Тайный Санта

Самостоятельный сайт на ASP.NET Core / Blazor с Identity/cookie и PostgreSQL.
Независимая C++20 библиотека с YASL проводит равномерную жеребьёвку.
Самодарение запрещено; исключения пар симметричны; остальные взаимные подарки
разрешены. Каждый дарит и получает ровно один подарок.

## Быстрый запуск в Docker

Нужны Git, Docker и Docker Compose. Submodule должен быть загружен до сборки.

```bash
git submodule update --init third_party/yasl
cp .env.example .env
# Замените POSTGRES_PASSWORD в .env своим случайным паролем.
docker compose build
docker compose up -d db
docker compose run --rm santa --migrate
docker compose up -d santa
```

Открыть http://localhost:5080. Этот compose предназначен для локального HTTP
и использует Development. Для публичного размещения см. [эксплуатацию](docs/operations.md).
Миграции запускаются отдельной командой, а не каждым web-процессом при старте.

1. Организатор и участники регистрируются по email и паролю.
2. Организатор создаёт игру, добавляет email зарегистрированных участников.
   Собственный email нужно добавить, если организатор тоже участвует.
3. Вводит пары исключений: два email через запятую в каждой строке.
4. Сохраняет состав и запускает жеребьёвку. После успеха состав неизменяем.
5. Каждый открывает игру под своим аккаунтом и видит только своего получателя.
6. Организатор завершает игру, закрывая доступ к результатам.

Письма, подтверждение email, гостевые приглашения и восстановление пароля
в MVP не реализованы. Email — идентификатор аккаунта, а не подтверждение владения
почтовым ящиком. Жеребьёвка с большим количеством строгих запретов может исчерпать
бюджет; сайт сообщает об этом и сохраняет Draft.

## Сборка без Docker (Linux)

Нужны C++20 compiler, CMake ≥ 3.20, .NET SDK 10 и PostgreSQL.

```bash
git submodule update --init third_party/yasl
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build -j 2
ctest --test-dir build --output-on-failure
dotnet restore santa/santa.csproj --locked-mode
export ConnectionStrings__Santa='Host=localhost;Database=santa;Username=santa;Password=YOUR_PASSWORD'
dotnet run --project santa -- --migrate
ASPNETCORE_ENVIRONMENT=Development dotnet run --project santa --no-launch-profile --urls http://localhost:5080
```

MSBuild копирует `build/libsanta_native.so` к серверной сборке и в publish.
Сначала собирайте native library. Для C++ API и алгоритма:
[core/README.md](core/README.md). Существующий второй шаблон `web/` сохранён;
работающий продукт расположен в `santa/`.

## Проверки

```bash
# Только отдельная тестовая БД: harness очищает её таблицы.
export SANTA_TEST_DB='Host=localhost;Database=santa_test;Username=santa;Password=YOUR_PASSWORD'
bash scripts/check.sh
# Против запущенного сайта: создаёт уникальные тестовые аккаунты и игры.
SANTA_TEST_URL=http://localhost:5080 python3 scripts/http-smoke.py
```

Тесты проверяют полный перебор малых графов, равномерный выбор, C ABI и Guid,
PostgreSQL-гонки, идемпотентность, устаревший снимок, откат частичной записи,
вход и права доступа. [Результаты проверок](docs/validation.md).

[Архитектура](docs/design.md) · [Чек-лист](docs/implementation-plan.md) ·
[Контекст продолжения](memory/project.md) · [Эксплуатация](docs/operations.md)

## Go-сервис

В каталоге `go/` находится параллельная реализация сервера с теми же PostgreSQL-таблицами и native C++ ABI. Для локального запуска нужен Go с CGO и собранный `build/libsanta_native.so`. Переменные `DATABASE_URL` или `ConnectionStrings__DefaultConnection` задают подключение. Миграция Go-сессий применяется командой `santa-go --migrate`. Docker-образ собирается из корня через `go/Dockerfile`, сервис `santa-go` описан в `compose.yaml`.
