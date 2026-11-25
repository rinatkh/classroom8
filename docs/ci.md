# CI — Classroom 8

Workflow: [.github/workflows/ci.yml](../.github/workflows/ci.yml). Он запускается на push, pull request (включая PR из fork) и вручную через Actions → Classroom 8 CI → Run workflow. Репозиторий должен содержать `go.mod` и Makefile в корне. Go берётся из `go.mod` (1.22), автоматическая смена toolchain отключена.

Локальная проверка перед PR:

```bash
make ci
make package
make bench-smoke
```

`make ci` выполняет deps/mod-check, компиляцию, gofmt-check, vet, unit/integration, race, порог покрытия 90% и build. `make package` упаковывает проверенные бинарники. `bench-smoke` вызывает все 112 сценариев по одному разу без порога скорости. Длинные benchmark не входят в обязательную проверку.

В Actions откройте неуспешный job и первый красный step: лог показывает команду, пакет и subtest. Coverage, отчёты smoke и архив Linux-бинарников сохраняются как artifact на 7 дней, если соответствующие файлы успели появиться. Они доступны на странице конкретного run; результаты smoke нельзя интерпретировать как производительность.

Workflow имеет только `contents: read`, не использует secrets, `pull_request_target` или публикацию в чужие сервисы. Деплой здесь не нужен: распространяемый результат — проверенные учебные исполняемые файлы. Настройка required checks в защите ветки делается владельцем репозитория. Push/PR запускают проверки автоматически, когда Actions разрешены в настройках репозитория; для PR из fork может потребоваться стандартное одобрение владельца.

Если работаете в macOS, для локального `make ci` нужен установленный Go и C toolchain для race (`xcode-select --install`, если его ещё нет). Workflow использует Ubuntu. Обновлять минимальную версию Go ради CI не требуется.
