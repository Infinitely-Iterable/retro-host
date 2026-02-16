document.addEventListener('DOMContentLoaded', async () => {
    const params = new URLSearchParams(window.location.search);
    const system = params.get('system');
    const rom = params.get('rom');
    const core = params.get('core');

    if (!system || !rom || !core) {
        document.getElementById('game-title').textContent = 'Missing parameters';
        return;
    }

    const romName = rom.replace(/\.[^.]+$/, '');
    document.getElementById('game-title').textContent = romName;
    document.title = `RetroHost - ${romName}`;

    // EmulatorJS configuration — must be set BEFORE loading loader.js
    window.EJS_player = '#game';
    window.EJS_gameUrl = `/roms/${system}/${rom}`;
    window.EJS_core = core;
    window.EJS_pathtodata = '/emulatorjs/';
    window.EJS_startOnLoaded = true;
    window.EJS_RESET_THREADING = true;
    window.EJS_threads = false;

    // Try to load existing save state
    try {
        const saveRes = await fetch(`/api/saves/${system}/${encodeURIComponent(romName)}`);
        if (saveRes.ok) {
            const saveData = await saveRes.arrayBuffer();
            window.EJS_loadStateURL = URL.createObjectURL(new Blob([saveData]));
        }
    } catch (e) {
        // No save found, that's fine
    }

    // Load EmulatorJS
    const script = document.createElement('script');
    script.src = '/emulatorjs/loader.js';
    document.body.appendChild(script);

    // Save button
    document.getElementById('btn-save').addEventListener('click', async () => {
        await saveState();
    });

    // Load button
    document.getElementById('btn-load').addEventListener('click', async () => {
        await loadState();
    });

    async function saveState() {
        const btn = document.getElementById('btn-save');
        try {
            if (!window.EJS_emulator) {
                btn.textContent = 'Not ready';
                setTimeout(() => btn.textContent = 'Save', 2000);
                return;
            }

            // EmulatorJS v4 save state API
            const saveData = window.EJS_emulator.gameManager.getSaveFile();
            if (!saveData) {
                btn.textContent = 'No data';
                setTimeout(() => btn.textContent = 'Save', 2000);
                return;
            }

            btn.textContent = 'Saving...';
            const res = await fetch(`/api/saves/${system}/${encodeURIComponent(romName)}`, {
                method: 'POST',
                body: saveData,
            });

            btn.textContent = res.ok ? 'Saved!' : 'Error';
        } catch (e) {
            console.error('Save failed:', e);
            btn.textContent = 'Error';
        }
        setTimeout(() => btn.textContent = 'Save', 2000);
    }

    async function loadState() {
        const btn = document.getElementById('btn-load');
        try {
            btn.textContent = 'Loading...';
            const res = await fetch(`/api/saves/${system}/${encodeURIComponent(romName)}`);
            if (!res.ok) {
                btn.textContent = 'No save';
                setTimeout(() => btn.textContent = 'Load', 2000);
                return;
            }

            const saveData = await res.arrayBuffer();
            if (window.EJS_emulator) {
                window.EJS_emulator.gameManager.loadSaveFile(new Uint8Array(saveData));
                btn.textContent = 'Loaded!';
            } else {
                btn.textContent = 'Not ready';
            }
        } catch (e) {
            console.error('Load failed:', e);
            btn.textContent = 'Error';
        }
        setTimeout(() => btn.textContent = 'Load', 2000);
    }

    // Auto-save on page unload
    window.addEventListener('beforeunload', () => {
        if (window.EJS_emulator) {
            try {
                const saveData = window.EJS_emulator.gameManager.getSaveFile();
                if (saveData) {
                    navigator.sendBeacon(
                        `/api/saves/${system}/${encodeURIComponent(romName)}`,
                        saveData
                    );
                }
            } catch (e) {
                // Best effort
            }
        }
    });
});
