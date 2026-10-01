# Trill Telegram Cloudflare Worker

This Worker is the first production slice of the Trill Telegram Core pipeline:

```text
Telegram
  -> Cloudflare Worker
  -> Neon Postgres
```

For this milestone, the Worker only receives, verifies, normalizes, and persists Telegram messages. It does **not** call the Trill agent yet.

## What it stores

The Worker writes to the existing Neon tables:

- `telegram_users`
- `telegram_chats`
- `telegram_messages`

Duplicate Telegram deliveries are safe because `telegram_messages` uses the existing unique constraint on `(telegram_chat_id, telegram_message_id)`.

## Required secrets

From `worker/`:

```bash
npx wrangler secret put DATABASE_URL
npx wrangler secret put TELEGRAM_WEBHOOK_SECRET
```

Keep `TELEGRAM_BOT_TOKEN` out of source control. It is not required by the Worker until the reply loop is added.

## Local setup

```bash
cd worker
npm install
npm test
npm run dev
```

Health check:

```bash
curl http://localhost:8787/health
```

## Deploy

```bash
npm run deploy
```

Wrangler will print the deployed Worker URL, for example:

```text
https://trill-telegram-webhook.<your-subdomain>.workers.dev
```

## Register the Telegram webhook

Generate a long random webhook secret and store the exact same value in the Worker as `TELEGRAM_WEBHOOK_SECRET`.

Then register the Worker endpoint with Telegram:

```bash
curl -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook" \
  -H "content-type: application/json" \
  -d '{
    "url": "https://YOUR-WORKER.workers.dev/telegram/webhook",
    "secret_token": "YOUR_TELEGRAM_WEBHOOK_SECRET",
    "allowed_updates": ["message", "edited_message"],
    "drop_pending_updates": false
  }'
```

Telegram will include the secret in the `X-Telegram-Bot-Api-Secret-Token` header. The Worker rejects requests that do not match it.

## Verify the first real message

1. Send a message to the bot in Telegram.
2. Check Worker logs:

```bash
npx wrangler tail
```

3. In Neon, confirm the rows exist:

```sql
SELECT * FROM telegram_users ORDER BY updated_at DESC LIMIT 5;
SELECT * FROM telegram_chats ORDER BY updated_at DESC LIMIT 5;
SELECT * FROM telegram_messages ORDER BY created_at DESC LIMIT 10;
```

Expected result: one Telegram message is persisted end-to-end.

## Next slice

Only after the persistence test passes:

```text
Telegram
  -> Worker
  -> Neon
  -> agent_sessions
  -> Trill Astro Buzz
  -> Telegram sendMessage
```

Build small. Compose big.
