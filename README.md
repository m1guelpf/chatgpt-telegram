# ChatGPT-bot

> Interact with ChatGPT

Go CLI to fuels a Telegram bot that lets you interact with [ChatGPT](https://openai.com/blog/chatgpt/), a large language model trained by OpenAI.

## Installation

Download the file corresponding to your OS in the [releases page](https://github.com/m1guelpf/chatgpt-telegram/releases/latest). After you extract it, copy `env.example` to `.env` and fill in your Bot's details (you'll need your bot token, which you can find [here](https://core.telegram.org/bots/tutorial#obtain-your-bot-token), and optionally your telegram id, which you can find by DMing `@userinfobot` on Telegram.

## Usage

Run the `chatgpt-telegram` binary!

## Browserless Authentication

By default, the program will launch a browser for you to sign into your account. If for whatever reason this isn't possible (compatibility issues, running on a server without a screen, etc.), you can manually provide your cookie.

To do this, first sign into ChatGPT on your browser, then open the Developer Tools, go to the Cookies section in the Application tab, and copy the value of the `__Secure-next-auth.session-token` cookie. Then, create a JSON file in your config dir (`/Users/[username]/Library/Application Support/chatgpt.json` in macOS, `C:\Users\[username]\AppData\Roaming\chatgpt.json` in Windows, `~/.config/chatgpt.json` in Linux), and write your cookie in the following format: `{ "openaisession": "YOUR_COOKIE_HERE" }`.

## License

This repository is licensed under the [MIT License](LICENSE).
