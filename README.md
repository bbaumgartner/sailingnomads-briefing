# Sailing Nomads Daily Briefing

A daily briefing generator for a sailing couple and their dog at https://sailingnomads.ch. It uses OpenAI with web search to produce a morning briefing covering weather, marine conditions, local events, news, sightseeing, and day planning — written directly into a Logseq journal.

## How it works

1. The **shell script** (`generate-briefing.sh`) reads the Logseq saillog to find the current GPS position and recent journal context
2. The **Go program** takes position and context as input, fetches weather/marine data, calls OpenAI with web search, and outputs a Logseq-formatted briefing to stdout
3. The shell script writes the briefing into today's journal file

## Requirements

- Go 1.22+
- OpenAI API key with access to `gpt-5` and web search

## Setup

```bash
# Clone and enter the repository
cd sailingnomads-briefing

# Build the Go program
go build -o briefing .

# Set your OpenAI API key
export OPENAI_API_KEY='sk-...'

# Copy and edit the config
cp config.env.example config.env
```

## GPS Position

The script reads the GPS position from your Logseq journal. Add a `current_position::` property to any recent journal entry:

```markdown
- current_position:: 47.13826/8.60032
```

The script scans from today backward (up to 30 days) and uses the first position it finds.

## Usage

### Manual run

```bash
./generate-briefing.sh /path/to/saillog ./config.env
```

### Go program directly

```bash
echo "some context" | go run . --lat 43.296 --lon 5.369 --lang de --prompt prompt.md
```

### Flags

| Flag       | Required | Default     | Description                        |
|------------|----------|-------------|------------------------------------|
| `--lat`    | yes      |             | Latitude                           |
| `--lon`    | yes      |             | Longitude                          |
| `--lang`   | no       | `de`        | Briefing language (de, en, fr, ..) |
| `--prompt` | no       | `prompt.md` | Path to the system prompt file     |

## Cron setup

To generate a briefing every morning at 06:00:

```bash
# Edit crontab
crontab -e

# Add this line (adjust paths)
0 6 * * * OPENAI_API_KEY='sk-...' /path/to/generate-briefing.sh /path/to/saillog /path/to/config.env >> /tmp/briefing.log 2>&1
```

## Customizing the prompt

The system prompt is in `prompt.md`. Edit it to change the briefing structure, tone, or sections. Changes take effect on the next run — no recompilation needed.

## Architecture

```
generate-briefing.sh          Go program (stdin → stdout)
┌─────────────────────┐       ┌──────────────────────────┐
│ Read config.env     │       │ Parse --lat, --lon       │
│ Find GPS position   │       │ Load prompt.md           │
│ Extract briefings   │──────>│ Reverse geocode          │
│ Extract logbook     │ stdin │ Fetch weather + marine   │
│ Build context       │       │ Call OpenAI + web search │
│                     │<──────│ Output markdown          │
│ Write to journal    │stdout │                          │
└─────────────────────┘       └──────────────────────────┘
```

## Dependencies

- [openai-go](https://github.com/openai/openai-go) — OpenAI API client
- [Open-Meteo](https://open-meteo.com/) — Weather and marine data (free, no key)
- [Nominatim](https://nominatim.openstreetmap.org/) — Reverse geocoding (free, no key)
