/* ── Topbar profile dropdown ── */

function toggleTopbarMenu() {
    const menu    = document.getElementById('topbarMenu');
    const profile = document.getElementById('topbarProfile');
    if (!menu) return;
    const open = menu.hidden;
    menu.hidden = !open;
    profile.classList.toggle('open', open);
    if (open) {
        setTimeout(() => document.addEventListener('click', closeTopbarMenuOnOutside), 0);
    } else {
        document.removeEventListener('click', closeTopbarMenuOnOutside);
    }
}

function closeTopbarMenuOnOutside(e) {
    const menu    = document.getElementById('topbarMenu');
    const profile = document.getElementById('topbarProfile');
    if (menu && !profile.contains(e.target)) {
        menu.hidden = true;
        profile.classList.remove('open');
        document.removeEventListener('click', closeTopbarMenuOnOutside);
    }
}

/* ── Theme ── */

function applyTheme(theme) {
    const resolved = theme === 'system'
        ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
        : theme;
    document.documentElement.setAttribute('data-theme', resolved);
}

function setTheme(theme) {
    localStorage.setItem('platform_theme', theme);
    applyTheme(theme);
    document.querySelectorAll('.st-theme-opt').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.theme === theme);
    });
}

function initTheme() {
    const saved = localStorage.getItem('platform_theme') || 'light';
    applyTheme(saved);
    document.querySelectorAll('.st-theme-opt').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.theme === saved);
    });
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        if (localStorage.getItem('platform_theme') === 'system') applyTheme('system');
    });
}

/* ── Sidebar: desktop collapse / mobile drawer ── */

function isMobile() {
    return window.innerWidth <= 768;
}

function toggleSidebar() {
    if (isMobile()) {
        const sidebar = document.getElementById('sidebar');
        if (sidebar.classList.contains('mobile-open')) {
            closeMobileDrawer();
        } else {
            openMobileDrawer();
        }
    } else {
        const layout  = document.querySelector('.ds-layout');
        const sidebar = document.getElementById('sidebar');
        const collapsed = sidebar.classList.toggle('collapsed');
        layout.classList.toggle('sidebar-collapsed', collapsed);
        localStorage.setItem('sidebar_collapsed', collapsed ? '1' : '0');
    }
}

function openMobileDrawer() {
    const sidebar  = document.getElementById('sidebar');
    const overlay  = document.getElementById('drawerOverlay');
    sidebar.classList.add('mobile-open');
    if (overlay) overlay.style.display = 'block';
    document.body.style.overflow = 'hidden';
}

function closeMobileDrawer() {
    const sidebar  = document.getElementById('sidebar');
    const overlay  = document.getElementById('drawerOverlay');
    sidebar.classList.remove('mobile-open');
    if (overlay) overlay.style.display = '';
    document.body.style.overflow = '';
}

/* Close drawer when a nav link is tapped on mobile */
document.addEventListener('click', function(e) {
    if (!isMobile()) return;
    const sidebar = document.getElementById('sidebar');
    if (!sidebar || !sidebar.classList.contains('mobile-open')) return;
    const link = e.target.closest('.ds-nav__item, .ds-back-link, .ws-pro-item, .ws-nav-leaf');
    if (link) closeMobileDrawer();
});

/* Close drawer on Escape */
document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') closeMobileDrawer();
});

/* On resize to desktop, reset mobile drawer state */
window.addEventListener('resize', function() {
    if (!isMobile()) closeMobileDrawer();
});

/* ── Init ── */

document.addEventListener('DOMContentLoaded', () => {
    if (!isMobile() && localStorage.getItem('sidebar_collapsed') === '1') {
        document.querySelector('.ds-layout').classList.add('sidebar-collapsed');
        document.getElementById('sidebar').classList.add('collapsed');
    }
    initTheme();
});

/* ── Login page submit animation ── */

const loginForm = document.querySelector('.auth-form');
if (loginForm) {
    loginForm.addEventListener('submit', () => {
        const btn = loginForm.querySelector('.btn--primary');
        if (btn) {
            btn.disabled = true;
            const t = btn.querySelector('.btn-text');
            if (t) t.textContent = 'Signing in\u2026';
        }
    });
}

/* ──────────────────────────────────────────────────────────────
   Universal mock-action handler
   Catches clicks on buttons that don't otherwise do anything,
   and surfaces a toast or modal so every control feels alive.
   ────────────────────────────────────────────────────────────── */
(function () {
    if (window.__mockBound) return;
    window.__mockBound = true;

    function ensureToast() {
        let t = document.getElementById('__mockToast');
        if (t) return t;
        t = document.createElement('div');
        t.id = '__mockToast';
        t.style.cssText = 'position:fixed;bottom:24px;right:24px;z-index:9999;background:#1a1e2e;color:#eef0f6;padding:14px 20px;border-radius:10px;font:400 13px/1.4 Inter,-apple-system,sans-serif;border:1px solid rgba(255,255,255,.12);box-shadow:0 10px 40px rgba(0,0,0,.4);display:flex;gap:10px;align-items:center;min-width:260px;max-width:440px;transform:translateY(20px);opacity:0;transition:all .2s;pointer-events:none';
        document.body.appendChild(t);
        return t;
    }

    function toast(msg, icon) {
        const t = ensureToast();
        t.innerHTML = '<span style="color:#10b981;font-size:16px;line-height:1">' + (icon || '\u2713') + '</span><div style="line-height:1.45">' + msg + '</div>';
        requestAnimationFrame(() => { t.style.transform = 'translateY(0)'; t.style.opacity = '1'; });
        clearTimeout(window.__mockTT);
        window.__mockTT = setTimeout(() => { t.style.transform = 'translateY(20px)'; t.style.opacity = '0'; }, 2600);
    }

    function ensureModal() {
        let m = document.getElementById('__mockModal');
        if (m) return m;
        m = document.createElement('div');
        m.id = '__mockModal';
        m.style.cssText = 'position:fixed;inset:0;z-index:10000;background:rgba(8,10,18,.55);backdrop-filter:blur(6px);display:none;align-items:center;justify-content:center;padding:20px';
        m.addEventListener('click', e => { if (e.target === m) close(); });
        document.addEventListener('keydown', e => { if (e.key === 'Escape') close(); });
        m.innerHTML =
            '<div id="__mockCard" style="background:#fff;color:#111;border-radius:14px;max-width:540px;width:100%;padding:0;box-shadow:0 30px 80px rgba(0,0,0,.45);border:1px solid rgba(0,0,0,.08);overflow:hidden;transform:scale(.97);transition:transform .18s">' +
              '<div style="padding:20px 24px 12px 24px;border-bottom:1px solid rgba(0,0,0,.06);display:flex;align-items:center;justify-content:space-between"><div id="__mockT" style="font:500 17px/1.3 Inter,sans-serif;letter-spacing:-.01em"></div>' +
              '<button id="__mockX" style="width:28px;height:28px;border-radius:7px;border:none;background:transparent;cursor:pointer;color:#9ca3af;font:500 18px/1 Inter,sans-serif">\u00d7</button></div>' +
              '<div id="__mockB" style="padding:20px 24px;font:300 14px/1.6 Inter,sans-serif;color:#4b5563;white-space:pre-wrap"></div>' +
              '<div id="__mockF" style="padding:14px 24px;border-top:1px solid rgba(0,0,0,.06);display:flex;gap:8px;justify-content:flex-end;background:#fafafa"></div>' +
            '</div>';
        document.body.appendChild(m);
        document.getElementById('__mockX').addEventListener('click', close);
        return m;
    }
    function applyModalTheme() {
        const c = document.getElementById('__mockCard');
        if (!c) return;
        const dark = document.documentElement.getAttribute('data-theme') === 'dark';
        if (dark) {
            c.style.background = '#1a1e2e';
            c.style.color = '#eef0f6';
            c.style.borderColor = 'rgba(255,255,255,.12)';
            document.getElementById('__mockB').style.color = '#9aa0b4';
            document.getElementById('__mockF').style.background = '#111421';
            document.getElementById('__mockF').style.borderColor = 'rgba(255,255,255,.08)';
        } else {
            c.style.background = '#fff';
            c.style.color = '#111';
            c.style.borderColor = 'rgba(0,0,0,.08)';
            document.getElementById('__mockB').style.color = '#4b5563';
            document.getElementById('__mockF').style.background = '#fafafa';
            document.getElementById('__mockF').style.borderColor = 'rgba(0,0,0,.06)';
        }
    }

    function close() {
        const m = document.getElementById('__mockModal');
        if (m) m.style.display = 'none';
    }

    function modal(title, body, buttons) {
        const m = ensureModal();
        document.getElementById('__mockT').textContent = title;
        document.getElementById('__mockB').textContent = body;
        const f = document.getElementById('__mockF');
        f.innerHTML = '';
        (buttons || [{ label: 'Got it', primary: true }]).forEach(b => {
            const btn = document.createElement('button');
            btn.textContent = b.label;
            btn.style.cssText = 'padding:9px 18px;border-radius:8px;font:400 13px Inter,sans-serif;cursor:pointer;border:1px solid transparent;transition:all .15s;' +
                (b.primary ? 'background:#7c5cff;color:#fff' : b.danger ? 'background:rgba(239,68,68,.1);color:#ef4444;border-color:rgba(239,68,68,.25)' : 'background:transparent;color:#6b7280;border-color:rgba(0,0,0,.1)');
            btn.addEventListener('click', () => {
                close();
                if (b.onClick) b.onClick();
                else if (b.toast) toast(b.toast);
            });
            f.appendChild(btn);
        });
        applyModalTheme();
        m.style.display = 'flex';
        requestAnimationFrame(() => { document.getElementById('__mockCard').style.transform = 'scale(1)'; });
    }

    window.mockToast = toast;
    window.mockModal = modal;
    window.mockConfirm = function (title, body, confirmLabel, onConfirm) {
        modal(title, body, [
            { label: 'Cancel' },
            { label: confirmLabel || 'Confirm', primary: true, onClick: onConfirm }
        ]);
    };

    // ── Replace native window.alert / window.confirm with the styled modal ──
    // window.confirm returns a boolean synchronously which we can't fully
    // emulate with an async modal, but we intercept inline `onsubmit="return
    // confirm(...)"` and `onclick="return confirm(...)"` patterns to render
    // the styled modal and submit the form on confirmation.
    window.alert = function (message) {
        modal('Notice', String(message || ''), [{ label: 'OK', primary: true }]);
    };

    // Intercept forms with onsubmit="return confirm('…')" — defer submission,
    // show our styled modal, submit programmatically on confirm.
    document.addEventListener('submit', function (e) {
        const f = e.target;
        if (!f || f.__mockConfirmed) return;
        const handler = f.getAttribute('onsubmit') || '';
        const m = handler.match(/return\s+confirm\(['"]([^'"]+)['"]\)/);
        if (!m) return;
        e.preventDefault();
        const msg = m[1];
        modal('Confirm', msg, [
            { label: 'Cancel' },
            { label: 'Confirm', primary: true, onClick: function () {
                f.__mockConfirmed = true;
                f.removeAttribute('onsubmit'); // prevent native confirm fallback
                if (typeof f.requestSubmit === 'function') f.requestSubmit();
                else f.submit();
            }}
        ]);
    }, true);

    // Intercept buttons with onclick="return confirm('…')" — same pattern,
    // re-trigger the originally-intended click on confirm.
    document.addEventListener('click', function (e) {
        const el = e.target.closest('button, a');
        if (!el || el.__mockConfirmed) return;
        const handler = el.getAttribute('onclick') || '';
        const m = handler.match(/return\s+confirm\(['"]([^'"]+)['"]\)/);
        if (!m) return;
        e.preventDefault();
        e.stopPropagation();
        const msg = m[1];
        modal('Confirm', msg, [
            { label: 'Cancel' },
            { label: 'Confirm', primary: true, onClick: function () {
                el.__mockConfirmed = true;
                el.removeAttribute('onclick');
                el.click(); // re-fire — now without our interception
            }}
        ]);
    }, true);

    // Click handler — intercept buttons that don't already have real behavior.
    document.addEventListener('click', function (e) {
        const el = e.target.closest('button, a.btn, [role="button"]');
        if (!el) return;
        if (el.__mockHandled) return;

        // Skip if button already has a real handler / href
        if (el.hasAttribute('onclick')) return;
        if (el.tagName === 'A' && el.getAttribute('href') && el.getAttribute('href') !== '#') return;
        if (el.type === 'submit') return;
        if (el.closest('form') && !el.type) return;
        if (el.closest('.theme-toggle')) return;
        if (el.closest('.ws-sidebar') || el.closest('.ds-sidebar') || el.closest('.sa-sidebar')) return;
        if (el.closest('[data-no-mock]')) return;

        // Identify the button's purpose from its text/label
        const txt = (el.textContent || el.getAttribute('aria-label') || 'Action').trim().replace(/\s+/g, ' ').slice(0, 80);
        if (!txt || txt.length < 2) return;

        // Determine intent by keywords
        const t = txt.toLowerCase();
        e.preventDefault();

        if (/^(cancel|dismiss|close|back)$/.test(t)) { close(); return; }

        if (/delete|destroy|remove|terminate|drop/.test(t)) {
            mockConfirm(
                txt,
                'This would permanently remove the resource. In this demo environment, no real change is made.',
                txt,
                function () { toast(txt + ' \u2014 done (demo)'); }
            );
            return;
        }

        if (/restart|reboot|reload|restart|restart/.test(t)) {
            toast(txt + ' triggered \u2014 takes ~30 s (demo)');
            return;
        }

        if (/save|apply|update|submit|confirm|approve/.test(t)) {
            toast('Saved \u2014 changes applied (demo)');
            return;
        }

        if (/download|export|csv|pdf|report/.test(t)) {
            toast(txt + ' \u2014 check your Downloads (demo)');
            return;
        }

        if (/new|create|add|launch|provision|deploy|install/.test(t)) {
            modal(txt, 'A creation wizard would walk you through sizing, region, networking, credentials, and tagging. In this demo, nothing is actually provisioned.',
                [{ label: 'Cancel' }, { label: 'Continue', primary: true, toast: 'Resource scheduled for creation (demo)' }]);
            return;
        }

        if (/view|open|details|manage|configure|settings|edit|logs|history/.test(t)) {
            modal(txt, 'This would open the detail view for this item \u2014 metrics, config, logs, and management actions.\n\nYou are in the demo environment.');
            return;
        }

        // Default
        toast(txt + ' \u2014 (demo action)');
    }, true);
})();

/* ── Session inactivity timeout ─────────────────────────────────────────────
   Warn after WARN_MS of no interaction; auto-logout after TIMEOUT_MS total.
   Only active on authenticated pages (those that include dashboard.js).
   ─────────────────────────────────────────────────────────────────────────── */
(function () {
    var WARN_MS    = 25 * 60 * 1000;  // 25 min  → show warning
    var TIMEOUT_MS = 30 * 60 * 1000;  // 30 min  → logout
    var warnTimer, logoutTimer, countdownInterval;

    function resetTimers() {
        clearTimeout(warnTimer);
        clearTimeout(logoutTimer);
        clearInterval(countdownInterval);
        hideWarning();
        warnTimer    = setTimeout(showWarning,  WARN_MS);
        logoutTimer  = setTimeout(doLogout,     TIMEOUT_MS);
    }

    function showWarning() {
        var el = document.getElementById('session-timeout-warning');
        if (!el) return;
        el.hidden = false;
        var remaining = 5 * 60; // seconds left
        var countEl = el.querySelector('.sto-countdown');
        function tick() {
            remaining--;
            if (countEl) countEl.textContent = Math.floor(remaining / 60) + ':' + ('0' + (remaining % 60)).slice(-2);
            if (remaining <= 0) clearInterval(countdownInterval);
        }
        tick();
        countdownInterval = setInterval(tick, 1000);
    }

    function hideWarning() {
        var el = document.getElementById('session-timeout-warning');
        if (el) el.hidden = true;
        clearInterval(countdownInterval);
    }

    function doLogout() {
        window.location.href = '/logout';
    }

    // Inject the warning banner into the DOM once
    document.addEventListener('DOMContentLoaded', function () {
        var banner = document.createElement('div');
        banner.id = 'session-timeout-warning';
        banner.hidden = true;
        banner.setAttribute('role', 'alert');
        banner.innerHTML =
            '<div class="sto-inner">' +
            '<svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.2"/><path d="M8 5v3.5l2.5 1.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>' +
            '<span>Your session will expire in <strong class="sto-countdown">5:00</strong> due to inactivity.</span>' +
            '<button class="btn btn--primary btn--sm" onclick="window._stoStayLoggedIn()">Stay logged in</button>' +
            '</div>';
        document.body.appendChild(banner);
        resetTimers();
    });

    window._stoStayLoggedIn = function () {
        resetTimers();
    };

    // Reset on any user interaction
    ['mousemove', 'keydown', 'mousedown', 'touchstart', 'scroll'].forEach(function (ev) {
        document.addEventListener(ev, function () {
            if (!document.getElementById('session-timeout-warning') ||
                document.getElementById('session-timeout-warning').hidden) {
                resetTimers();
            }
        }, { passive: true });
    });
})();
