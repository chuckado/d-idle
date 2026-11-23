// The alerts badge of the top band: how many alerts the notification centre holds, toned by the worst one, from GET
// /api/alerts. It is a link to the Alerts section of the Settings sheet (?settings=alerts, nav.js opens it); with no alert, or with the alerts route off, the badge is not shown.
export function tone(alerts) {
  return (alerts || []).some((a) => a.severity === 'alarm') ? 'bad' : 'warn';
}

// The page's own address with settings=alerts: the sheet opens on Alerts and the token and layout of the address stay.
export function settingsHref(search) {
  const kept = String(search || '').replace(/^\?/, '').split('&').filter((p) => p && !/^settings=/.test(p));
  kept.push('settings=alerts');
  return '?' + kept.join('&');
}

export function mount(root, ctx) {
  const a = ctx.el('a', 'd-badge');
  a.href = settingsHref(typeof location === 'object' && location ? location.search : '');
  a.dataset.key = 'alerts';
  root.appendChild(a);
  return {
    update(v) {
      const n = v && Array.isArray(v.alerts) ? v.alerts.length : 0;
      if (!n) { ctx.hide(true); return; }
      ctx.hide(false);
      a.textContent = 'alerts ' + n;
      a.dataset.tone = tone(v.alerts);
      a.title = v.alerts[n - 1].text || '';
    },
  };
}
