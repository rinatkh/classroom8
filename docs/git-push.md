# Git push: проверка результата и HTTP 400

Ошибка `RPC failed; HTTP 400` и `remote end hung up unexpectedly` относится к передаче Git по HTTP. Строка `Everything up-to-date` после fatal не подтверждает, что новый commit появился на сервере. Конкретную причину (прокси, HTTP-согласование, сеть или ответ сервера) один такой лог не устанавливает.

## 1. Проверить коммит на сервере

Из корня проекта:

```bash
git rev-parse HEAD
git ls-remote --exit-code origin refs/heads/main
```

Первый SHA — ваш локальный commit. Первый столбец второй команды — актуальный SHA main на origin. Равенство означает, что вершины совпадают. При ошибке второй команды проверка не состоялась. Локальная origin/main может быть устаревшей; её одной недостаточно.

Для текущей ветки это же делает `make git-check-push`: выводит два SHA и возвращает 0 только при точном равенстве. Команда не изменяет файлы/refs и ничего не публикует. Несовпадение вершин не определяет, кто впереди: после `git fetch origin` посмотрите `git log --oneline --left-right HEAD...origin/main`.

## 2. Повторить HTTPS push с HTTP/1.1

Если локальный commit ещё не отправлен, для main:

```bash
git -c http.version=HTTP/1.1 push origin main
make git-check-push
```

Это разовая настройка команды. Она проверяет гипотезу о проблеме HTTP-транспорта и не гарантирует устранение любого HTTP 400. После успешного push убедитесь в совпадении SHA.

В исходном сообщении pack был 1,41 MiB. Такой объём сам по себе не означает слишком большой Go-проект. Увеличение http.postBuffer меняет работу chunked POST, но официальная документация Git не считает его общим решением push-ошибок. Если подтверждено, что промежуточный прокси не поддерживает chunked, можно отдельно проверить небольшой буфер на один push:

```bash
git -c http.version=HTTP/1.1 -c http.postBuffer=4194304 push origin main
```

Это обход для конкретной несовместимости, а не доказанная причина текущей ошибки. Если сбой остаётся, проверьте соединение с GitHub/настройки своей сети и при уже настроенном SSH используйте доступ через SSH по официальной инструкции GitHub. Настройки HTTPS-прокси/ключей не входят в исходники Go и не исправляются заменой handlers или алгоритмов.

## 3. Проверки и публикация новых изменений

Для classroom перед коммитом: `make ci`. Для стартового homework с TODO: `make starter-check`; функциональные `make test` должны выявлять нереализованные задания.

```bash
git status --short
git add -A
git commit -m "Fix lesson 8 project files"
git -c http.version=HTTP/1.1 push origin main
make git-check-push
```

Если коммитить нечего, новый commit создавать не требуется. GitHub Actions запускается после того, как push действительно обновил серверную ветку; локальная проверка не является удалённым CI-run.

Источники: [Git http.version/http.postBuffer](https://git-scm.com/docs/git-config), [Git ls-remote](https://git-scm.com/docs/git-ls-remote), [GitHub connectivity](https://docs.github.com/en/get-started/using-github/troubleshooting-connectivity-problems).
