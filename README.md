# ChatGPT-bot

> Interact with ChatGPT

Go CLI to fuels a Telegram bot that lets you interact with [ChatGPT](https://openai.com/blog/chatgpt/), a large language model trained by OpenAI.

## Installation

Download the file corresponding to your OS in the [releases page](https://github.com/m1guelpf/chatgpt-telegram/releases/latest). After you extract it, copy `env.example` to `.env` and fill in your Bot's details (you'll need your bot token, which you can find [here](https://core.telegram.org/bots/tutorial#obtain-your-bot-token), and optionally your telegram id, which you can find by DMing @userinfobot on Telegram.

## Usage

- Run the `chatgpt-telegram` binary
- Message the bot with the `/chatgpt` prefix:
```
/chatgpt hello
> Hello! How can I help you today? Is there something specific you'd like to know or talk about? I'm a large language model trained by OpenAI and I'm here to assist you with any questions you might have.
```

## License

This repository is licensed under the [MIT License](LICENSE).
