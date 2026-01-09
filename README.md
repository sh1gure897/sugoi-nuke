# sugoi-nuke

Nuke

## 機能
- **Nuke**: `!r` で起動
- **AllBan**: `!ban` で起動

## 使用方法

### 1. ビルド
Go 1.20以上が必要です。
```bash
git clone https://github.com/sh1gure897/sugoi-nuke.git
cd sugoi-nuke
go mod tidy
go build -ldflags="-s -w" -o renewer main.go
