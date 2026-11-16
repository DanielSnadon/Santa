# Тайный Санта

Небольшой самостоятельный сайт для жеребьёвки Тайного Санты.

Основное приложение находится в `santa/`: ASP.NET Core, Blazor и PostgreSQL.
Жеребьёвка выполняется нативным C++20 модулем через C ABI. YASL подключается
как Git submodule в `third_party/yasl`.

Правила простые:

- каждый участник дарит одному человеку и получает подарок от одного человека;
- себя выбрать нельзя;
- исключение `A — B` действует в обе стороны;
- взаимное дарение для остальных пар разрешено.

## Запуск через Docker

Нужны Git, Docker и Docker Compose.

```bash
git submodule update --init third_party/yasl
cp .env.example .env
# Задайте в .env случайное значение POSTGRES_PASSWORD.
docker compose build
docker compose up -d db
docker compose run --rm santa --migrate
docker compose up -d santa
```

Сайт будет доступен по адресу <http://localhost:5080>.

Миграция запускается отдельной командой. Сервис не меняет схему базы данных при
каждом старте.

### Как пользоваться

1. Зарегистрируйте аккаунты участников.
2. Создайте игру и добавьте email участников. Если организатор участвует,
   добавьте и его email.
3. При необходимости укажите пары исключений, по одной паре в строке.
4. Сохраните состав и запустите жеребьёвку.
5. Участник может оставить своему Санте сообщение с пожеланиями. В сообщении
   поддерживается безопасный Markdown.
6. После жеребьёвки каждый участник увидит только своего получателя и его
   сообщение для Санты.
7. После закрытия игры результаты больше не открываются.

В текущей версии нет писем, подтверждения email, гостевых приглашений и
восстановления пароля. Email используется как имя аккаунта.

## Сборка без Docker

Нужны Linux, C++20 compiler, CMake 3.20 или новее, .NET SDK 10 и PostgreSQL.

```bash
git submodule update --init third_party/yasl
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build -j 2
ctest --test-dir build --output-on-failure
dotnet restore santa/santa.csproj --locked-mode
export ConnectionStrings__Santa='Host=localhost;Database=santa;Username=santa;Password=YOUR_PASSWORD'
dotnet run --project santa -- --migrate
ASPNETCORE_ENVIRONMENT=Development dotnet run \
  --project santa --no-launch-profile --urls http://localhost:5080
```

Сначала собирается `build/libsanta_native.so`, затем .NET-проект копирует его в
серверную сборку и в publish.

## Проверки

Для полного локального прогона нужна отдельная тестовая база PostgreSQL:

```bash
export SANTA_TEST_DB='Host=localhost;Database=santa_test;Username=santa;Password=YOUR_PASSWORD'
bash scripts/check.sh
```

HTTP smoke запускается против уже работающего сайта:

```bash
SANTA_TEST_URL=http://localhost:5080 python3 scripts/http-smoke.py
```

Проверки включают C++ core, C ABI, отображение Guid, права доступа, безопасный
Markdown сообщений, гонки при публикации, повторные запросы и откат частичной
записи.

## Что где находится

- `santa/` — основной сайт;
- `core/` — C++ engine и C ABI;
- `third_party/yasl/` — исходник YASL;
- `scripts/` — smoke-проверки;
- `tests/` — интеграционный harness;
- `go/` — сохранённая параллельная реализация сервера;
- `web/` — второй шаблон, сохранённый для справки.
