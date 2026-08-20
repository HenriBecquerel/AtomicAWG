# Сборка AtomicAWG

Репозиторий состоит из двух независимых Go-модулей:

- **`GoApp/`** — само приложение (интерфейс на Fyne + движок, вкомпилированный
  напрямую, без отдельного исполняемого файла).
- **`Core/`** — движок туннеля/прокси; собирается и как библиотека (её
  использует `GoApp`), и как самостоятельный CLI `Core/cmd/wireproxy`.

Если нужен просто готовый бинарник — не собирайте вручную: при пуше в `main`
или тега `v*` GitHub Actions (`.github/workflows/build.yml`) сам собирает
все четыре таргета (macOS arm64/x64, Linux x64, Windows x64) и прикладывает
их к [Actions](../../actions) / [Releases](../../releases). Инструкция ниже —
для локальной сборки и разработки.

## Предварительные требования

- **Go** — версия, указанная в `GoApp/go.mod` (сейчас 1.26). Проверить:
  `go version`.
- **Компилятор C** — интерфейс использует CGO (нативные биндинги
  Cocoa/OpenGL/X11 через Fyne), поэтому `CGO_ENABLED=1` обязателен для сборки
  `GoApp`. Для `Core` (движок и CLI) CGO не нужен, там `CGO_ENABLED=0`.

### macOS

- Xcode Command Line Tools: `xcode-select --install`.
- Ad-hoc code signing использует системный `codesign` — он уже есть в CLT.

### Linux

Заголовки X11/OpenGL для GLFW-бэкенда Fyne:

```sh
# Debian/Ubuntu
sudo apt-get install gcc libgl1-mesa-dev xorg-dev libxcursor-dev \
  libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev

# Fedora
sudo dnf install gcc mesa-libGL-devel libXcursor-devel libXrandr-devel \
  libXinerama-devel libXi-devel libXxf86vm-devel

# Arch
sudo pacman -S gcc mesa libxcursor libxrandr libxinerama libxi libxxf86vm
```

### Windows

Нужен GCC-совместимый компилятор в `PATH` — проще всего через
[MSYS2](https://www.msys2.org/) (`pacman -S mingw-w64-ucrt-x86_64-gcc`) или
[TDM-GCC](https://jmeubank.github.io/tdm-gcc/). Стандартный MSVC (`cl.exe`)
для CGO/cgo-glfw не подходит.

## Сборка GoApp (приложение)

```sh
cd GoApp
go mod download
```

**macOS** — собирает подписанный `.app`-бандл (трей, без иконки в Dock) и zip:

```sh
./Build-Mac.sh arm64   # или: ./Build-Mac.sh x64
# результат: dist/macos-<arch>/AtomicAWG.app и dist/AtomicAWG-macOS-<arch>.zip
```

**Windows** — собирает `.exe` с флагом `-H=windowsgui` (без чёрного окна
консоли) и иконкой (см. ниже):

```powershell
./Build-Windows.ps1
# результат: dist/AtomicAWG.exe
```

**Linux** — отдельного скрипта упаковки пока нет, достаточно обычной сборки:

```sh
go build -trimpath -ldflags "-X main.appVersion=dev" -o dist/AtomicAWG .
```

### Иконка exe (Windows)

Значок встраивается через объектный файл `GoApp/rsrc_windows_amd64.syso`,
который уже лежит в репозитории — линковщик Go подхватывает его
автоматически, ничего дополнительно делать не нужно. Пересоздать его нужно
только если меняется сама иконка (`assets/atomicawg.ico`):

```sh
go run github.com/tc-hib/go-winres@latest simply --icon assets/atomicawg.ico --manifest gui --out rsrc_windows
mv rsrc_windows_windows_amd64.syso rsrc_windows_amd64.syso
rm -f rsrc_windows_windows_386.syso
```

Запуск без сборки (для разработки):

```sh
go run .
```

## Сборка Core (движок и CLI)

```sh
cd Core
go mod download
make            # собирает бинарник wireproxy в Core/
# или напрямую:
CGO_ENABLED=0 go build -trimpath -o wireproxy ./cmd/wireproxy
```

## Тесты и проверки

Перед коммитом стоит прогонять то же, что делает CI:

```sh
cd Core  && go vet ./... && go test ./...
cd GoApp && go vet ./... && go test ./...
```

`gofmt -l .` в обоих каталогах должен ничего не выводить (значит, всё
отформатировано).

## Частые проблемы

- **`cgo: C compiler "gcc" not found`** — не установлен/не в `PATH`
  компилятор C (см. требования по платформе выше). Проверить: `gcc --version`
  или `clang --version`.
- **На macOS `codesign` ругается `resource fork ... not allowed`** — это
  случается, если каталог проекта лежит внутри синхронизируемой iCloud
  Drive-папки: macOS на лету помечает файлы `com.apple.FinderInfo`, что
  ломает сборку подписи. `Build-Mac.sh` уже собирает и подписывает во
  временном каталоге вне iCloud и копирует готовый бандл обратно — если
  ошибка всё же появляется, попробуйте собрать из каталога вне iCloud Drive.
- **Приложение не видит `Core`** — `GoApp/go.mod` подключает движок через
  `replace github.com/awgproxy/awgproxy => ../Core`, то есть `Core/` должен
  лежать рядом с `GoApp/` (как в этом репозитории), а не быть скачан
  отдельно.
