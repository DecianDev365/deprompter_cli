# Deprompt

A fast, minimal CLI tool that reverse engineers AI image generation prompts from any image. Drop in an image, get back a detailed prompt you can paste straight into Midjourney, DALL·E, Stable Diffusion, or Flux.

> Want a web UI instead? Try **[deprompter.pages.dev](https://deprompter.pages.dev)**

---

## Demo

```
$ deprompt image.png

  ─────────────────────────────────────────
  ⠋ Analyzing image.png
  ─────────────────────────────────────────

  ✓ Prompt generated

  "A smiling cartoon sun with a yellow face and rays on a dark blue background, simple vector graphic style, centered composition, bright and cheerful, smiling sun with eyebrows and winking, flat design, modern and minimalistic illustration."
  --ar 2:3 --v 6 --style raw

  ─────────────────────────────────────────
  ✓ Copied to clipboard
  ─────────────────────────────────────────
```

---

## Features

- Supports **Groq** and **Google Gemini** as AI providers
- Interactive setup wizard on first run — no flags needed after that
- Arrow key menu for selecting your provider
- Config saved locally at `~/.deprompt/config.json`
- Detailed prompts covering subject, style, lighting, composition, color, mood, and technical details
- Output automatically copied to clipboard
- Works with PNG, JPG, JPEG, and WEBP images

---

## Installation

### From source

Make sure you have [Go 1.21+](https://go.dev/dl/) installed.

```bash
git clone https://github.com/DecianDev365/deprompt-cli.git
cd deprompt-cli
go install .
```

After this, `deprompt` is available globally from anywhere in your terminal.

---

## Setup

On first run, deprompt will automatically launch an interactive setup wizard:

```bash
deprompt image.png
```

Or run setup manually anytime:

```bash
deprompt config
```

The wizard will ask you to:
1. Select your AI provider (Groq or Google Gemini) using arrow keys
2. Paste your API key

Your config is saved to `~/.deprompt/config.json` and never asked for again.

---

## Getting an API Key

**Groq** — free tier, 1500 requests/day
1. Go to [console.groq.com](https://console.groq.com)
2. Sign in → API Keys → Create API Key

**Google Gemini** — free tier, 1500 requests/day
1. Go to [aistudio.google.com](https://aistudio.google.com)
2. Sign in → Get API Key

---

## Usage

```bash
# basic usage — uses saved config
deprompt image.png

# override provider for a single run
deprompt --provider gemini image.png

# override provider and key for a single run
deprompt --provider groq --key YOUR_KEY image.png

# reconfigure provider and key
deprompt config
```

### Supported image formats

| Format | Extension |
|--------|-----------|
| PNG    | `.png`    |
| JPEG   | `.jpg`, `.jpeg` |
| WebP   | `.webp`   |

---

## How it works

```
deprompt image.png
       ↓
reads and converts image to base64
       ↓
sends to Groq or Gemini vision API
       ↓
AI analyzes subject, style, lighting,
composition, color, mood, and details
       ↓
returns a detailed, copy-ready prompt
```

---

## Configuration

Config is stored at `~/.deprompt/config.json`:

```json
{
  "provider": "groq",
  "key": "your-api-key"
}
```

You can edit this file directly or run `deprompt config` to update it interactively.

---

## Web Version

Prefer a visual interface? The web version of DePrompter is free and requires no installation.

**[deprompter.pages.dev](https://deprompter.pages.dev)**

Same functionality, same AI providers, runs entirely in your browser. Your API key is stored locally and never sent to any server other than the AI provider directly.

---

## Contributing

Contributions are welcome. Feel free to open issues or pull requests for bug fixes, new features, or improvements.  :)

```bash
git clone https://github.com/DecianDev365/deprompt-cli.git
cd deprompt-cli
go run main.go api.go image.go config.go ui.go -- --provider groq --key YOUR_KEY image.png
```

---

## License

MIT — see [LICENSE](LICENSE) for details.

---

<p align="center">Built by <a href="https://github.com/DecianDev365">DecianDev365</a></p>
