#!/bin/bash

# Скрипт для создания релиза PDF Compressor в Gitea
# Использование: ./scripts/release.sh v1.0.0

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Использование: $0 <version>"
    echo "Пример: $0 v1.0.0"
    exit 1
fi

echo "🚀 Создание релиза $VERSION для PDF Compressor"

# Проверяем что мы в правильной директории
if [ ! -f "go.mod" ]; then
    echo "❌ Ошибка: Запустите скрипт из корня проекта"
    exit 1
fi

# Проверяем что все изменения закоммичены
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ Ошибка: Есть незакоммиченные изменения"
    git status
    exit 1
fi

# Запускаем тесты
echo "🧪 Запуск тестов..."
go test ./... || {
    echo "❌ Тесты не прошли"
    exit 1
}

# Собираем для разных платформ
echo "🔨 Сборка бинарников..."
mkdir -p releases

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o releases/pdf-compressor-${VERSION}-windows-amd64.exe cmd/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o releases/pdf-compressor-${VERSION}-linux-amd64 cmd/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o releases/pdf-compressor-${VERSION}-darwin-amd64 cmd/main.go

# ARM64 versions
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o releases/pdf-compressor-${VERSION}-linux-arm64 cmd/main.go
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o releases/pdf-compressor-${VERSION}-darwin-arm64 cmd/main.go

# Создаем архивы
echo "📦 Создание архивов..."
cd releases

# Windows
zip pdf-compressor-${VERSION}-windows-amd64.zip pdf-compressor-${VERSION}-windows-amd64.exe
rm pdf-compressor-${VERSION}-windows-amd64.exe

# Linux
tar -czf pdf-compressor-${VERSION}-linux-amd64.tar.gz pdf-compressor-${VERSION}-linux-amd64
rm pdf-compressor-${VERSION}-linux-amd64

# macOS
tar -czf pdf-compressor-${VERSION}-darwin-amd64.tar.gz pdf-compressor-${VERSION}-darwin-amd64
rm pdf-compressor-${VERSION}-darwin-amd64

# ARM64
tar -czf pdf-compressor-${VERSION}-linux-arm64.tar.gz pdf-compressor-${VERSION}-linux-arm64
rm pdf-compressor-${VERSION}-linux-arm64

tar -czf pdf-compressor-${VERSION}-darwin-arm64.tar.gz pdf-compressor-${VERSION}-darwin-arm64
rm pdf-compressor-${VERSION}-darwin-arm64

cd ..

# Создаем и пушим тег
echo "🏷️ Создание тега..."
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"

echo "✅ Релиз подготовлен!"
echo "📁 Файлы релиза находятся в папке releases/"
echo "🌐 Теперь создайте релиз в Gitea веб-интерфейсе:"
echo "   1. Перейдите в ваш репозиторий в Gitea"
echo "   2. Нажмите 'Releases' → 'New Release'"
echo "   3. Выберите тег: $VERSION"
echo "   4. Загрузите файлы из папки releases/"
