document.addEventListener('DOMContentLoaded', async () => {
    const tabsEl = document.getElementById('system-tabs');
    const listEl = document.getElementById('rom-list');

    let systems = [];
    let activeSystem = null;

    try {
        const res = await fetch('/api/systems');
        systems = await res.json();
    } catch (err) {
        listEl.innerHTML = '<p class="empty-state">Failed to load systems. Is the server running?</p>';
        return;
    }

    if (!systems || systems.length === 0) {
        listEl.innerHTML = '<p class="empty-state">No ROMs found. Mount your ROM directory to /roms.</p>';
        return;
    }

    // Render system tabs
    systems.forEach(sys => {
        const tab = document.createElement('button');
        tab.className = 'system-tab';
        tab.dataset.system = sys.id;
        tab.innerHTML = `
            <span class="system-dot ${sys.id}"></span>
            ${sys.name}
            <span class="count">${sys.romCount}</span>
        `;
        tab.addEventListener('click', () => selectSystem(sys));
        tabsEl.appendChild(tab);
    });

    // Auto-select first system
    selectSystem(systems[0]);

    async function selectSystem(sys) {
        activeSystem = sys;

        // Update tab active states
        tabsEl.querySelectorAll('.system-tab').forEach(tab => {
            tab.classList.toggle('active', tab.dataset.system === sys.id);
        });

        // Fetch ROMs
        try {
            const res = await fetch(`/api/roms?system=${sys.id}`);
            const roms = await res.json();
            renderROMs(roms, sys);
        } catch (err) {
            listEl.innerHTML = '<p class="empty-state">Failed to load ROMs.</p>';
        }
    }

    function renderROMs(roms, sys) {
        if (!roms || roms.length === 0) {
            listEl.innerHTML = '<p class="empty-state">No ROMs for this system.</p>';
            return;
        }

        roms.sort((a, b) => a.name.localeCompare(b.name));

        // Group ROMs by tag
        const groups = {};
        roms.forEach(rom => {
            const tag = rom.tag || '';
            if (!groups[tag]) groups[tag] = [];
            groups[tag].push(rom);
        });

        // Sort group names, untagged last
        const tagNames = Object.keys(groups).sort((a, b) => {
            if (a === '') return 1;
            if (b === '') return -1;
            return a.localeCompare(b);
        });

        listEl.innerHTML = '';

        tagNames.forEach(tag => {
            if (tag && tagNames.length > 1) {
                const heading = document.createElement('h2');
                heading.className = 'tag-heading';
                heading.textContent = tag;
                listEl.appendChild(heading);
            } else if (!tag && tagNames.length > 1) {
                const heading = document.createElement('h2');
                heading.className = 'tag-heading';
                heading.textContent = 'Untagged';
                listEl.appendChild(heading);
            }

            const grid = document.createElement('div');
            grid.className = 'rom-grid';

            groups[tag].forEach(rom => {
                const card = document.createElement('a');
                card.className = 'rom-card';
                card.href = `/player.html?system=${sys.id}&rom=${encodeURIComponent(rom.fileName)}&core=${sys.core}`;
                card.innerHTML = `
                    <div class="rom-name">${escapeHtml(rom.name)}</div>
                    <div class="rom-file">${escapeHtml(rom.fileName)}</div>
                    ${rom.tag ? `<div class="rom-tag">${escapeHtml(rom.tag)}</div>` : ''}
                `;
                grid.appendChild(card);
            });

            listEl.appendChild(grid);
        });
    }

    function escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
});
