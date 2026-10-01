import { describe, expect, it } from "vitest";
import { extractTelegramMessage, safeEqual } from "../src/telegram.js";

describe("safeEqual", () => {
  it("accepts identical values", () => {
    expect(safeEqual("relay713", "relay713")).toBe(true);
  });

  it("rejects different values", () => {
    expect(safeEqual("relay713", "relay714")).toBe(false);
    expect(safeEqual("short", "longer")).toBe(false);
  });
});

describe("extractTelegramMessage", () => {
  it("normalizes a Telegram message update", () => {
    const update = {
      update_id: 42,
      message: {
        message_id: 7,
        from: {
          id: 28133,
          username: "trill",
          first_name: "Trill",
          language_code: "en"
        },
        chat: {
          id: 713,
          type: "private"
        },
        text: "Explorer online."
      }
    };

    expect(extractTelegramMessage(update)).toMatchObject({
      updateId: 42,
      messageId: 7,
      chatId: 713,
      chatType: "private",
      userId: 28133,
      username: "trill",
      text: "Explorer online."
    });
  });

  it("ignores unsupported updates", () => {
    expect(extractTelegramMessage({ update_id: 1 })).toBeNull();
  });
});
