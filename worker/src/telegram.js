export function extractTelegramMessage(update) {
  const message = update?.message ?? update?.edited_message ?? null;
  if (!message?.chat?.id || !message?.message_id) return null;

  const from = message.from ?? null;

  return {
    updateId: update.update_id ?? null,
    messageId: message.message_id,
    chatId: message.chat.id,
    chatType: message.chat.type ?? null,
    chatTitle: message.chat.title ?? null,
    userId: from?.id ?? null,
    username: from?.username ?? null,
    firstName: from?.first_name ?? null,
    lastName: from?.last_name ?? null,
    languageCode: from?.language_code ?? null,
    text: message.text ?? message.caption ?? null,
    rawUpdate: update
  };
}

export function safeEqual(left, right) {
  if (typeof left !== "string" || typeof right !== "string") return false;

  const length = Math.max(left.length, right.length);
  let diff = left.length ^ right.length;

  for (let i = 0; i < length; i += 1) {
    diff |= (left.charCodeAt(i) || 0) ^ (right.charCodeAt(i) || 0);
  }

  return diff === 0;
}
