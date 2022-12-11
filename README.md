# ChatGPT-bot

> Interact with ChatGPT

Go CLI to fuels a Telegram bot that lets you interact with [ChatGPT](https://openai.com/blog/chatgpt/), a large language model trained by OpenAI.

## Installation
Download the file corresponding to your OS in the [releases page](https://github.com/m1guelpf/chatgpt-telegram/releases/latest). 
- `chatgpt-telegram-Darwin-amd64`: macOS (Intel)
- `chatgpt-telegram-Darwin-arm64`: macOS (M1)
- `chatgpt-telegram-Linux-amd64`: Linux
- `chatgpt-telegram-Linux-arm64`: Linux (ARM)
- `chatgpt-telegram-Win-amd64`: Windows

After you download the file, extract it into a folder and open the `env.example` file with a text editor and fill in your credentials. 
- `TELEGRAM_TOKEN`: Your Telegram Bot token
  - Follow [this guide](https://core.telegram.org/bots/tutorial#obtain-your-bot-token) to create a bot and get the token.
- `TELEGRAM_ID` (Optional): Your Telegram User ID
  - If you set this, only you will be able to interact with the bot.
  - To get your ID, message `@userinfobot` on Telegram.
  - Multiple IDs can be provided, separated by commas.
- `EDIT_WAIT_SECONDS` (Optional): Amount of seconds to wait between edits
  - This is set to `1` by default, but you can increase if you start getting a lot of `Too Many Requests` errors.
- Save the file, and rename it to `.env`.
> **Note** Make sure you rename the file to _exactly_ `.env`! The program won't work otherwise.

Finally, open the terminal in your computer (if you're on windows, look for `PowerShell`), navigate to the path you extracted the above file (you can use `cd dirname` to navigate to a directory, ask ChatGPT if you need more assistance 😉) and run `./chatgpt-telegram`.

### Running with Docker

If you're trying to run this on a server with an existing Docker setup, you might want to use our Docker image instead.

```sh
docker pull ghcr.io/m1guelpf/chatgpt-telegram
```

Here's how you'd set things up with `docker-compose`:

```yaml
services:
  chatgpt-telegram:
    image: ghcr.io/m1guelpf/chatgpt-telegram
    container_name: chatgpt-telegram
    volumes:
      # your ".config" local folder must include a "chatgpt.json" file
      - .config/:/root/.config
    environment:
      - TELEGRAM_ID=
      - TELEGRAM_TOKEN=
```

> **Note** The docker setup is optimized for the Browserless authentication mechanism, described below. Make sure you update the `.config/chatgpt.json` file in this repo with your session token before running.

## Authentication

By default, the program will launch a browser for you to sign into your account, and close it once you're signed in. If this setup doesn't work for you (there are issues with the browser starting, you want to run this in a computer with no screen, etc.), you can manually extract your session from your browser instead.

To do this, first sign in to ChatGPT on your browser, then open the Developer Tools (right click anywhere in the page, then click "Inspect"), click on the Application tab and then on the Cookies section, and copy the value of the `__Secure-next-auth.session-token` cookie.

You will then have to create a config file in the following location depending on your OS (replace `YOUR_USERNAME_HERE` with your username:

- `~/.config/chatgpt.json`: Linux
- `C:\Users\YOUR_USERNAME_HERE\AppData\Roaming\chatgpt.json`: Windows
- `/Users/YOUR_USERNAME_HERE/Library/Application Support/chatgpt.json`: macOS

> **Note** If you have already run the program, the file should exist but be empty. If it doesn't exist yet, you can either run the program or manually create it.

Finally, add your cookie to the file and save it. It should look like this: `{ "openaisession": "YOUR_COOKIE_HERE" }`.

## License

This repository is licensed under the [MIT License](LICENSE).
