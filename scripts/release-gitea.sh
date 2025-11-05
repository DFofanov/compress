#!/bin/bash
# Скрипт автоматической генерации релиза на Gitea для PDF Compressor
# Автор: PDF Compressor Team
# Версия: 1.0.0

set -e  # Остановка при ошибках

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Переменные конфигурации
BINARY_NAME="pdf-compressor"
BUILD_DIR="releases"
GITEA_SERVER=""  # Заполните URL вашего Gitea сервера
GITEA_TOKEN=""   # Заполните токен доступа Gitea
GITEA_OWNER=""   # Заполните владельца репозитория
GITEA_REPO="pdf-compressor"

# Функция вывода сообщений
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Функция проверки зависимостей
check_dependencies() {
    log "Проверка зависимостей..."
    
    # Проверяем Go
    if ! command -v go &> /dev/null; then
        error "Go не установлен"
    fi
    
    # Проверяем git
    if ! command -v git &> /dev/null; then
        error "Git не установлен"
    fi
    
    # Проверяем curl для API запросов
    if ! command -v curl &> /dev/null; then
        error "curl не установлен"
    fi
    
    # Проверяем zip
    if ! command -v zip &> /dev/null; then
        error "zip не установлен"
    fi
    
    log "Все зависимости найдены"
}

# Функция проверки переменных окружения
check_env() {
    log "Проверка переменных окружения..."
    
    if [ -z "$GITEA_SERVER" ]; then
        error "GITEA_SERVER не установлен"
    fi
    
    if [ -z "$GITEA_TOKEN" ]; then
        error "GITEA_TOKEN не установлен"
    fi
    
    if [ -z "$GITEA_OWNER" ]; then
        error "GITEA_OWNER не установлен"
    fi
    
    log "Переменные окружения проверены"
}

# Функция получения версии
get_version() {
    if [ -n "$1" ]; then
        VERSION="$1"
    elif [ -f "VERSION" ]; then
        VERSION=$(cat VERSION)
    else
        VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
    fi
    
    # Добавляем префикс v если его нет
    if [[ ! $VERSION =~ ^v ]]; then
        VERSION="v$VERSION"
    fi
    
    log "Версия релиза: $VERSION"
}

# Функция проверки git статуса
check_git_status() {
    log "Проверка состояния git..."
    
    # Проверяем что мы в git репозитории
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        error "Не найден git репозиторий"
    fi
    
    # Проверяем что нет незафиксированных изменений
    if ! git diff-index --quiet HEAD --; then
        warn "Есть незафиксированные изменения"
        read -p "Продолжить? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Проверяем что мы на master/main ветке
    CURRENT_BRANCH=$(git branch --show-current)
    if [[ "$CURRENT_BRANCH" != "master" && "$CURRENT_BRANCH" != "main" ]]; then
        warn "Вы не на master/main ветке (текущая: $CURRENT_BRANCH)"
        read -p "Продолжить? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Функция запуска тестов
run_tests() {
    log "Запуск тестов..."
    
    if ! go test ./...; then
        error "Тесты не прошли"
    fi
    
    log "Все тесты прошли успешно"
}

# Функция создания тега
create_tag() {
    log "Создание тега $VERSION..."
    
    # Проверяем что тег еще не существует
    if git rev-parse "$VERSION" >/dev/null 2>&1; then
        warn "Тег $VERSION уже существует"
        read -p "Перезаписать? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git tag -d "$VERSION"
        else
            exit 1
        fi
    fi
    
    # Создаем аннотированный тег
    RELEASE_NOTES="Release $VERSION

✨ Новые возможности:
- Обновления и улучшения интерфейса
- Оптимизация производительности

🐛 Исправления:
- Различные багфиксы и улучшения стабильности

📦 Поддерживаемые платформы:
- Windows (64-bit)
- Linux (64-bit, ARM64)
- macOS (Intel 64-bit, Apple Silicon ARM64)"
    
    git tag -a "$VERSION" -m "$RELEASE_NOTES"
    
    # Отправляем тег в origin
    git push origin "$VERSION"
    
    log "Тег $VERSION создан и отправлен"
}

# Функция сборки бинарников
build_binaries() {
    log "Сборка бинарников для разных платформ..."
    
    # Создаем директорию для релиза
    RELEASE_DIR="$BUILD_DIR/$VERSION"
    mkdir -p "$RELEASE_DIR"
    
    # Массив платформ
    platforms=(
        "windows/amd64"
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
    )
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -r GOOS GOARCH <<< "$platform"
        output="$RELEASE_DIR/${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}"
        
        if [ "$GOOS" = "windows" ]; then
            output="${output}.exe"
        fi
        
        log "Сборка для $GOOS/$GOARCH"
        
        # Сборка с флагами оптимизации
        GOOS=$GOOS GOARCH=$GOARCH go build \
            -ldflags="-s -w -X main.version=$VERSION -X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
            -o "$output" \
            cmd/main.go
        
        if [ $? -eq 0 ]; then
            log "✅ $GOOS/$GOARCH построен успешно"
        else
            error "❌ Ошибка сборки для $GOOS/$GOARCH"
        fi
    done
}

# Функция создания архивов
create_archives() {
    log "Создание архивов..."
    
    cd "$BUILD_DIR/$VERSION"
    
    # Windows - ZIP архивы
    for file in *windows*.exe; do
        if [ -f "$file" ]; then
            archive="${file%.exe}.zip"
            zip "$archive" "$file"
            rm "$file"
            log "Создан архив: $archive"
        fi
    done
    
    # Linux и macOS - TAR.GZ архивы
    for file in *linux* *darwin*; do
        if [ -f "$file" ] && [[ ! "$file" == *.zip ]] && [[ ! "$file" == *.tar.gz ]]; then
            archive="${file}.tar.gz"
            tar -czf "$archive" "$file"
            rm "$file"
            log "Создан архив: $archive"
        fi
    done
    
    cd - > /dev/null
}

# Функция создания релиза в Gitea
create_gitea_release() {
    log "Создание релиза в Gitea..."
    
    # JSON для создания релиза
    RELEASE_JSON=$(cat <<EOF
{
  "tag_name": "$VERSION",
  "name": "PDF Compressor $VERSION",
  "body": "# 🔥 PDF Compressor $VERSION\n\n## ✨ Новые возможности\n- Обновления и улучшения\n- Оптимизация производительности\n\n## 🐛 Исправления\n- Различные багфиксы\n- Улучшения стабильности\n\n## 📦 Установка\n1. Скачайте архив для вашей платформы\n2. Распакуйте и запустите\n\n## 📖 Документация\nПолная документация доступна в README.md",
  "draft": false,
  "prerelease": false
}
EOF
    )
    
    # Создаем релиз через API
    RESPONSE=$(curl -s -X POST \
        -H "Authorization: token $GITEA_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$RELEASE_JSON" \
        "$GITEA_SERVER/api/v1/repos/$GITEA_OWNER/$GITEA_REPO/releases")
    
    # Получаем ID релиза
    RELEASE_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | cut -d':' -f2 | head -n1)
    
    if [ -z "$RELEASE_ID" ]; then
        error "Не удалось создать релиз. Ответ: $RESPONSE"
    fi
    
    log "Релиз создан с ID: $RELEASE_ID"
    
    # Загружаем архивы
    log "Загрузка архивов..."
    
    for archive in "$BUILD_DIR/$VERSION"/*; do
        if [ -f "$archive" ]; then
            filename=$(basename "$archive")
            log "Загрузка $filename..."
            
            curl -s -X POST \
                -H "Authorization: token $GITEA_TOKEN" \
                -F "attachment=@$archive" \
                "$GITEA_SERVER/api/v1/repos/$GITEA_OWNER/$GITEA_REPO/releases/$RELEASE_ID/assets"
            
            if [ $? -eq 0 ]; then
                log "✅ $filename загружен"
            else
                warn "❌ Ошибка загрузки $filename"
            fi
        fi
    done
}

# Функция очистки
cleanup() {
    log "Очистка временных файлов..."
    # Здесь можно добавить очистку при необходимости
}

# Функция показа справки
show_help() {
    echo -e "${BLUE}Скрипт генерации релиза PDF Compressor${NC}"
    echo ""
    echo "Использование: $0 [версия]"
    echo ""
    echo "Параметры:"
    echo "  версия    Версия релиза (например: v1.2.0)"
    echo "            Если не указана, используется VERSION файл или последний git тег"
    echo ""
    echo "Переменные окружения:"
    echo "  GITEA_SERVER    URL Gitea сервера"
    echo "  GITEA_TOKEN     Токен доступа Gitea"
    echo "  GITEA_OWNER     Владелец репозитория"
    echo ""
    echo "Примеры:"
    echo "  $0                    # Автоматическое определение версии"
    echo "  $0 v1.2.0            # Конкретная версия"
    echo ""
}

# Основная функция
main() {
    echo -e "${BLUE}🚀 PDF Compressor Release Generator${NC}"
    echo ""
    
    # Обработка аргументов
    case "${1:-}" in
        -h|--help|help)
            show_help
            exit 0
            ;;
    esac
    
    # Основной процесс
    check_dependencies
    check_env
    get_version "$1"
    check_git_status
    run_tests
    create_tag
    build_binaries
    create_archives
    create_gitea_release
    cleanup
    
    log "🎉 Релиз $VERSION успешно создан!"
    echo ""
    echo -e "${GREEN}Релиз доступен по адресу:${NC}"
    echo "$GITEA_SERVER/$GITEA_OWNER/$GITEA_REPO/releases/tag/$VERSION"
    echo ""
}

# Обработка сигналов
trap cleanup EXIT

# Запуск основной функции
main "$@"
