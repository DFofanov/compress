# Git Hooks for Gitea

Automatic synchronization with GitHub after push to Gitea.

## 📁 Files

- **post-receive** - Hook for automatic push to GitHub

## 🚀 Installation

### 1. Find Repository Path on Gitea Server

```bash
# Usually one of these paths:
/var/lib/gitea/data/gitea-repositories/username/compress.git
# or
/home/git/gitea-repositories/username/compress.git
```

### 2. Copy Hook

```bash
# On Gitea server
cd /path/to/gitea-repositories/username/compress.git/hooks/

# Copy file
cp /path/to/this/repo/gitea-hooks/post-receive ./

# Or download directly
wget https://raw.githubusercontent.com/username/compress/main/gitea-hooks/post-receive
```

### 3. Configure Hook

```bash
nano post-receive
```

Change:
```bash
GITHUB_REPO="git@github.com:YOUR_USERNAME/compress.git"
```

### 4. Make Executable

```bash
chmod +x post-receive
chown git:git post-receive
```

### 5. Create Log Directory

```bash
mkdir -p /var/log/gitea
chown git:git /var/log/gitea
```

## 🔑 Authentication Setup

### Option A: SSH (Recommended)

```bash
# On Gitea server as git user
sudo su - git
ssh-keygen -t ed25519 -C "gitea-sync"

# Copy public key
cat ~/.ssh/id_ed25519.pub

# Add to GitHub:
# Settings → SSH and GPG keys → New SSH key

# Verify
ssh -T git@github.com
```

### Option B: HTTPS with Token

1. Create Personal Access Token on GitHub
   - Settings → Developer settings → Personal access tokens
   - Scope: `repo`

2. Use in hook:
```bash
GITHUB_REPO="https://YOUR_TOKEN@github.com/username/compress.git"
```

## ✅ Verification

```bash
# Test push
cd /tmp
git clone http://gitea.example.com/username/compress.git
cd compress
echo "test" >> README.md
git add README.md
git commit -m "Test sync"
git push

# Check log
tail -f /var/log/gitea/github-sync.log

# Check GitHub - changes should appear in 1-2 seconds
```

## 📊 What Hook Does

1. ✅ Monitors pushes to `main` and `master` branches
2. ✅ Automatically pushes to GitHub
3. ✅ Synchronizes tags
4. ✅ Logs all operations
5. ✅ Shows beautiful output with emojis

## 🐛 Troubleshooting

### Hook Not Firing

```bash
# Check permissions
ls -la post-receive
# Should be: -rwxr-xr-x

# Check owner
chown git:git post-receive

# Check syntax
bash -n post-receive
```

### Permission Denied

```bash
# For SSH
ssh -T git@github.com

# Check key permissions
chmod 600 ~/.ssh/id_ed25519

# For HTTPS - check token
```

### Can't Find Git

```bash
# Add PATH to beginning of hook:
export PATH=/usr/bin:/usr/local/bin:$PATH
```

## 📝 Logs

```bash
# View sync logs
tail -f /var/log/gitea/github-sync.log

# Clear old logs
> /var/log/gitea/github-sync.log
```

## 🔄 Alternatives

If Git Hook doesn't work, see other methods in [GITEA_SYNC_EN.md](../GITEA_SYNC_EN.md):
- GitHub Actions (every hour)
- Gitea Mirror (built-in feature)
- Double remote (locally)

---

**See also**: [GITEA_SYNC_EN.md](../GITEA_SYNC_EN.md) for detailed documentation
