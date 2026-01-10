# sugoi-nuke

とにかく速さを求めたnuke

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
main.go
```
### 2. 超簡単なコンフィグ設定
```json
{
  "token": "BOT_TOKEN",
  "trigger": "!r",
  "banTrigger": "!ban",
  "serverName": "name",
  "webhookName": "name",
  "setup": {
    "channelName": "name",
    "channelCount": 25,
    "messageContent": "@everyone anal lol",
    "messageCount": 15
  }
}
```
# 免責事項
本ツールは教育的、または自身の管理下にあるサーバーのテスト目的で作成されています。
悪用や嫌がらせ目的での使用は絶対にしないでください。このツールを使用して発生したトラブルについて、作者（sh1gure897）は一切の責任を負いません。
