// The audit behind card uireason (docs/ui-design-system.md, "Reason lines"): a control that is off (disabled, or aria-disabled) and
// on screen must name a visible, non-empty reason line with aria-describedby, the line must sit beside it (inside the control's
// grandparent), and the title attribute is not the reason. unexplained(nodes) returns a text for each control that breaks that.
// nodes: every element of a page (a real DOM: [...root.querySelectorAll('*')]; the node tests' stub DOM: its flat list).
// Classic script: window.dWhy, or module.exports for the node tests.
(function (g) {
  'use strict';
  const off = (n) => !!n.disabled || (n.getAttribute && n.getAttribute('aria-disabled') === 'true');
  const label = (n) => (n.dataset && n.dataset.key) || n.id || n.tag || n.tagName || '?';
  function unexplained(nodes) {
    const out = [], parent = new Map(); // a stub DOM may not set parents: derive them from the children lists
    for (const n of nodes) for (const c of n.children || []) if (!parent.has(c)) parent.set(c, n);
    const up = (n) => (n && (n.parentNode || parent.get(n))) || null;
    const shown = (n) => { for (let x = n; x; x = up(x)) if (x.hidden) return false; return true; };
    for (const n of nodes) {
      const tag = String(n.tag || n.tagName || '').toLowerCase();
      if (!['button', 'input', 'select', 'textarea'].includes(tag) || !off(n) || !shown(n)) continue;
      const id = n.getAttribute('aria-describedby');
      const line = id && nodes.find((x) => x.id === id);
      if (!line || !shown(line) || !String(line.textContent || '').trim()) { out.push(label(n) + ': no visible reason line'); continue; }
      const gp = up(up(n));
      let inside = false;
      for (let x = line; x && !inside; x = up(x)) inside = x === gp;
      if (!inside) out.push(label(n) + ': its reason line is not beside it');
      if (n.getAttribute('title')) out.push(label(n) + ': a title attribute (not a reason)');
    }
    return out;
  }
  const api = { unexplained };
  if (typeof module !== 'undefined' && module.exports) module.exports = api;
  else g.dWhy = api;
})(typeof window !== 'undefined' ? window : globalThis);
