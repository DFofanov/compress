# Git Hooks для Gitea

Автоматическая синхронизация с GitHub после push в Gitea.

## 📁 Файлы

- **post-receive** - Hook для автоматического push в GitHub

## 🚀 Установка

### 1. Найдите путь к репозиторию на сервере Gitea

```bash
# Обычно это один из путей:
/var/lib/gitea/data/gitea-repositories/username/compress.git
# или
/home/git/gitea-repositories/username/compress.git
```

### 2. Скопируйте hook

```bash
# На сервере Gitea
cd /path/to/gitea-repositories/username/compress.git/hooks/

# Скопируйте файл
cp /path/to/this/repo/gitea-hooks/post-receive ./

# Или загрузите напрямую
wget https://raw.githubusercontent.com/username/compress/main/gitea-hooks/post-receive
```

### 3. Настройте hook

```bash
nano post-receive
```

Измените:
```bash
GITHUB_REPO="git@github.com:YOUR_USERNAME/compress.git"
```

### 4. Сделайте исполняемым

```bash
chmod +x post-receive
chown git:git post-receive
```

### 5. Создайте директорию для логов

```bash
mkdir -p /var/log/gitea
chown git:git /var/log/gitea
```

## 🔑 Настройка аутентификации

### Вариант A: SSH (Рекомендуется)

```bash
# На сервере Gitea под пользователем git
sudo su - git
ssh-keygen -t ed25519 -C "gitea-sync"

# Скопируйте публичный ключ
cat ~/.ssh/id_ed25519.pub

# Добавьте на GitHub:
# Settings → SSH and GPG keys → New SSH key

# Проверьте
ssh -T git@github.com
```

### Вариант B: HTTPS с токеном

1. Создайте Personal Access Token на GitHub
   - Settings → Developer settings → Personal access tokens
   - Scope: `repo`

2. Используйте в hook:
```bash
GITHUB_REPO="https://YOUR_TOKEN@github.com/username/compress.git"
```

## ✅ Проверка

```bash
# Тестовый push
cd /tmp
git clone http://gitea.example.com/username/compress.git
cd compress
echo "test" >> README.md
git add README.md
git commit -m "Test sync"
git push

# Проверьте лог
tail -f /var/log/gitea/github-sync.log

# Проверьте GitHub - изменения должны появиться через 1-2 секунды
```

## 📊 Что делает hook

1. ✅ Отслеживает push в ветки `main` и `master`
2. ✅ Автоматически пушит в GitHub
3. ✅ Синхронизирует теги
4. ✅ Логирует все операции
5. ✅ Показывает красивый вывод с эмодзи

## 🐛 Устранение проблем

### Hook не срабатывает

```bash
# Проверьте права
ls -la post-receive
# Должно быть: -rwxr-xr-x

# Проверьте владельца
chown git:git post-receive

# Проверьте синтаксис
bash -n post-receive
```

### Permission denied

```bash
# Для SSH
ssh -T git@github.com

# Проверьте права на ключ
chmod 600 ~/.ssh/id_ed25519

# Для HTTPS - проверьте токен
```

### Не находит git

```bash
# Добавьте PATH в начало hook:
export PATH=/usr/bin:/usr/local/bin:$PATH
```

## 📝 Логи

```bash
# Просмотр логов синхронизации
tail -f /var/log/gitea/github-sync.log

# Очистка старых логов
> /var/log/gitea/github-sync.log
```

## 🔄 Альтернативы

Если Git Hook не подходит, см. другие методы в [GITEA_SYNC.md](../GITEA_SYNC.md):
- GitHub Actions (каждый час)
- Gitea Mirror (встроенная функция)
- Двойной remote (локально)

---

**См. также**: [GITEA_SYNC.md](../GITEA_SYNC.md) для подробной документации

