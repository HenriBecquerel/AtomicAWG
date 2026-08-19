# AtomicAWG

Нативный десктопный клиент WireGuard и AmneziaWG с локальным SOCKS5-прокси.

Движок (форк [wireproxy](https://github.com/pufferffish/wireproxy) с поддержкой
[AmneziaWG](https://github.com/amnezia-vpn/amneziawg-go)) собран в тот же
бинарник, что и интерфейс — никакого второго исполняемого файла, который
извлекается на диск и запускается отдельным процессом.

## Структура репозитория

- **[`GoApp/`](GoApp)** — само приложение: интерфейс на [Fyne](https://fyne.io/),
  движок подключён напрямую как Go-библиотека. Живёт в трее, окно открывается
  по клику на иконку — как нативный клиент WireGuard.
- **[`Core/`](Core)** — движок туннеля и прокси-серверов (SOCKS5/HTTP/SNI) как
  отдельный Go-модуль; также собирается в самостоятельный CLI
  (`Core/cmd/wireproxy`) для серверного/безголового использования — см.
  [`Core/README.md`](Core/README.md).

## Сборка

Готовые сборки под все четыре платформы (macOS arm64/x64, Linux x64,
Windows x64) публикуются автоматически GitHub Actions при пуше тега —
смотрите вкладку [Actions](../../actions) и [Releases](../../releases).

Для локальной сборки — краткая версия:

```sh
cd GoApp && go build .        # запустить без упаковки
./Build-Mac.sh arm64          # или: собрать подписанный .app (только macOS)
```

Полная инструкция с требованиями по платформам и разбором частых проблем —
в [BUILDING.md](BUILDING.md).

## Лицензия

MIT, см. [LICENSE](LICENSE). Ядро AmneziaWG распространяется по собственной
лицензии — см. соответствующий upstream-репозиторий.

## Разработчик

Antoine Henri Becquerel — [github.com/HenriBecquerel](https://github.com/HenriBecquerel)
