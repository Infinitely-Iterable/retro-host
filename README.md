# RetroHost

A self-hosted, browser-based retro gaming platform. Point it at a directory of legally obtained ROMs and play them from any device on your network. Save states sync automatically so you can pick up where you left off on any device.

## Supported Systems

| System | File Extensions |
|--------|----------------|
| Game Boy | `.gb` |
| Game Boy Color | `.gbc` |
| Game Boy Advance | `.gba` |
| NES | `.nes` |
| SNES | `.smc`, `.sfc` |

## Quick Start

### Docker Compose (Recommended)

Create a `docker-compose.yml`:

```yaml
services:
  retro-host:
    image: ghcr.io/infinitely-iterable/retro-host:latest
    container_name: retro-host
    ports:
      - "6969:6969"
    volumes:
      - /path/to/your/roms:/roms:ro
      - /path/to/appdata/retro-host:/data
    environment:
      - PORT=6969
    restart: unless-stopped
```

Replace `/path/to/your/roms` with the directory containing your ROM files and start it:

```bash
docker compose pull
docker compose up -d
```

Open `http://your-server:6969` in a browser.

### Unraid

Use the following volume paths:

```yaml
volumes:
  - /mnt/user/data/roms:/roms:ro
  - /mnt/user/appdata/retro-host:/data
```

### Docker Run

```bash
docker run -d \
  --name retro-host \
  -p 6969:6969 \
  -v /path/to/your/roms:/roms:ro \
  -v /path/to/appdata/retro-host:/data \
  -e PORT=6969 \
  --restart unless-stopped \
  ghcr.io/infinitely-iterable/retro-host:latest
```

## ROM Directory

ROMs are detected by file extension. Your directory structure does not matter — RetroHost recursively scans all subdirectories. Organize however you like:

```
/roms/
├── gba/
│   └── pokemon.gba
├── snes/
│   └── zelda.smc
└── tetris.gb
```

Files with non-ROM extensions (`.srm`, `.sav`, `.bak`, `.ips`, `.ups`) are automatically filtered out.

If you place ROMs inside a subdirectory named after a system (`gb`, `gbc`, `gba`, `nes`, `snes`), files in that directory are assigned to that system regardless of extension.

## Tagging / Grouping ROMs

You can organize ROMs into visual groups by creating a `tags.json` file in your data directory (e.g., `/mnt/user/appdata/retro-host/tags.json`):

```json
{
    "Pokemon Fire Red.gba": "RPG",
    "Pokemon Emerald.gba": "RPG",
    "Super Mario World.smc": "Platformer",
    "Mega Man X.smc": "Platformer",
    "Zelda - A Link to the Past.smc": "Adventure"
}
```

Keys must match the **exact filename** including extension. Tagged ROMs are grouped under headings on the game selection screen. Untagged ROMs appear at the bottom.

## Save States

Save states are stored server-side in the data directory under `/data/saves/`. This means:

- Saves persist across container rebuilds
- Any device on your network shares the same saves
- You can back up saves by copying the data directory

Use the **Save** and **Load** buttons in the player toolbar, or saves will auto-sync when you close the tab.

## CLI Usage

You can list ROMs and generate player URLs by exec-ing into the container:

```bash
# List all detected ROMs
docker exec retro-host /app/retro-host list

# Get a URL to play a specific ROM (fuzzy match by name)
docker exec retro-host /app/retro-host play zelda

# Show help
docker exec retro-host /app/retro-host help
```

Set the `HOST_ADDR` environment variable so the `play` command generates correct URLs:

```yaml
environment:
  - HOST_ADDR=192.168.1.100:6969
```

## Keyboard Controls

| Action | Key |
|--------|-----|
| D-Pad | Arrow Keys |
| A | X |
| B | Z |
| X (SNES) | S |
| Y (SNES) | A |
| L | Q |
| R | W |
| Start | Enter |
| Select | Right Shift |

Controls can be remapped in the EmulatorJS settings menu (gear icon) during gameplay. Remappings persist in your browser's local storage.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ROM_DIR` | `/roms` | Path to ROM directory inside the container |
| `DATA_DIR` | `/data` | Path to persistent data (saves, tags.json) |
| `PORT` | `6969` | Server port |
| `HOST_ADDR` | `localhost:PORT` | External address used for CLI-generated URLs |

## Notes

### Vimium Users

If you use the Vimium browser extension, add your RetroHost URL (e.g., `http://192.168.1.100:6969/*`) to Vimium's **Excluded URLs** list. Vimium's keyboard shortcuts will intercept emulator input and make games unplayable. To exclude:

1. Click the Vimium icon → **Options**
2. Under **Excluded URLs and keys**, add: `http://your-server:6969/*`
3. Save changes

### Legal Disclaimer

RetroHost does not include or distribute any ROM files or copyrighted game data. You are responsible for ensuring that any ROMs used with this software were legally obtained. Only use ROMs that you have the legal right to play.
