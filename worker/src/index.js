import { neon } from "@neondatabase/serverless";
import { extractTelegramMessage, safeEqual } from "./telegram.js";

function json(data, status = 200) {
  return new Response(JSON.stringify(data), {
    status,
    headers: { "content-type": "application/json; charset=utf-8" }
  });
}

async function persistTelegramMessage(env, item) {
  if (!env.DATABASE_URL) {
    throw new Error("DATABASE_URL is not configured");
  }

  const sql = neon(env.DATABASE_URL);

  if (item.userId !== null) {
    await sql`
      INSERT INTO telegram_users (
        telegram_user_id,
        username,
        first_name,
        last_name,
        language_code,
        updated_at
      )
      VALUES (
        ${item.userId},
        ${item.username},
        ${item.firstName},
        ${item.lastName},
        ${item.languageCode},
        now()
      )
      ON CONFLICT (telegram_user_id)
      DO UPDATE SET
        username = EXCLUDED.username,
        first_name = EXCLUDED.first_name,
        last_name = EXCLUDED.last_name,
        language_code = EXCLUDED.language_code,
        updated_at = now()
    `;
  }

  await sql`
    INSERT INTO telegram_chats (
      telegram_chat_id,
      chat_type,
      title,
      updated_at
    )
    VALUES (
      ${item.chatId},
      ${item.chatType},
      ${item.chatTitle},
      now()
    )
    ON CONFLICT (telegram_chat_id)
    DO UPDATE SET
      chat_type = EXCLUDED.chat_type,
      title = EXCLUDED.title,
      updated_at = now()
  `;

  const inserted = await sql`
    INSERT INTO telegram_messages (
      telegram_message_id,
      telegram_chat_id,
      telegram_user_id,
      text,
      raw_update
    )
    VALUES (
      ${item.messageId},
      ${item.chatId},
      ${item.userId},
      ${item.text},
      ${JSON.stringify(item.rawUpdate)}::jsonb
    )
    ON CONFLICT (telegram_chat_id, telegram_message_id)
    DO NOTHING
    RETURNING id
  `;

  return {
    inserted: inserted.length === 1,
    recordId: inserted[0]?.id ?? null
  };
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);

    if (request.method === "GET" && url.pathname === "/health") {
      return json({
        ok: true,
        service: "trill-telegram-webhook",
        agentId: env.TRILL_AGENT_ID ?? "trill-astro-buzz"
      });
    }

    if (request.method !== "POST" || url.pathname !== "/telegram/webhook") {
      return json({ ok: false, error: "not_found" }, 404);
    }

    if (!env.TELEGRAM_WEBHOOK_SECRET) {
      console.error("TELEGRAM_WEBHOOK_SECRET is not configured");
      return json({ ok: false, error: "server_misconfigured" }, 500);
    }

    const receivedSecret =
      request.headers.get("X-Telegram-Bot-Api-Secret-Token") ?? "";

    if (!safeEqual(receivedSecret, env.TELEGRAM_WEBHOOK_SECRET)) {
      return json({ ok: false, error: "unauthorized" }, 401);
    }

    let update;
    try {
      update = await request.json();
    } catch {
      return json({ ok: false, error: "invalid_json" }, 400);
    }

    const message = extractTelegramMessage(update);

    if (!message) {
      return json({
        ok: true,
        ignored: true,
        reason: "unsupported_update"
      });
    }

    try {
      const result = await persistTelegramMessage(env, message);

      return json({
        ok: true,
        stored: result.inserted,
        duplicate: !result.inserted,
        recordId: result.recordId
      });
    } catch (error) {
      console.error("Failed to persist Telegram update", {
        message: error instanceof Error ? error.message : String(error),
        updateId: message.updateId,
        chatId: message.chatId,
        messageId: message.messageId
      });

      return json({ ok: false, error: "persistence_failed" }, 500);
    }
  }
};
