// A tiny DOM for the node tests of web/nav.js (stdlib only): nodes with a parent, events that bubble to the document, focus.
'use strict';
class Ev { constructor(type, init) { this.type = type; this.detail = init && init.detail; this.cancelable = !!(init && init.cancelable); this.defaultPrevented = false; this.target = null; }
  preventDefault() { if (this.cancelable) this.defaultPrevented = true; } stopPropagation() { this.stopped = true; } }
function makeDoc() {
  const doc = { readyState: 'complete', activeElement: null, listeners: {}, _ids: new Map() };
  class N {
    constructor(tag) { this.tag = tag; this.children = []; this.parent = null; this.className = ''; this.id = ''; this.dataset = {}; this.attrs = {}; this.hidden = false; this.textContent = ''; this.listeners = {}; this.href = ''; this.type = '';
      this.style = { props: {}, setProperty(k, v) { this.props[k] = v; } }; this.isConnected = true; }
    get firstChild() { return this.children[0] || null; }
    _detach(c) { if (c.parent) c.parent.children = c.parent.children.filter((x) => x !== c); c.parent = this; }
    appendChild(c) { this._detach(c); this.children.push(c); return c; }
    append(...cs) { cs.forEach((c) => this.appendChild(c)); }
    insertBefore(c, ref) { this._detach(c); const i = ref ? this.children.indexOf(ref) : -1; if (i < 0) this.children.push(c); else this.children.splice(i, 0, c); return c; }
    replaceChildren(...cs) { this.children.forEach((c) => { c.parent = null; }); this.children = []; cs.forEach((c) => this.appendChild(c)); }
    after(c) { const p = this.parent; p._detach(c); p.children.splice(p.children.indexOf(this) + 1, 0, c); }
    setAttribute(k, v) { this.attrs[k] = String(v); if (k === 'id') this.id = v; }
    getAttribute(k) { return this.attrs[k]; }
    removeAttribute(k) { delete this.attrs[k]; }
    addEventListener(t, f) { (this.listeners[t] = this.listeners[t] || []).push(f); }
    contains(n) { for (let x = n; x; x = x.parent) if (x === this) return true; return false; }
    focus() { doc.activeElement = this; }
    getBoundingClientRect() { return { top: 0, bottom: 40, left: 0, right: 100 }; }
    matches() { return false; }
    closest(sel) {
      const parts = sel.split(',').map((s) => s.trim());
      for (let x = this; x; x = x.parent) for (const p of parts) {
        const m = /^([a-z]+)(?::not\(\.([\w-]+)\))?$/.exec(p);
        if (m && x.tag === m[1] && !(m[2] && (' ' + x.className + ' ').includes(' ' + m[2] + ' '))) return x;
      }
      return null;
    }
    dispatchEvent(ev) { ev.target = ev.target || this; for (let x = this; x && !ev.stopped; x = x.parent) (x.listeners[ev.type] || []).slice().forEach((f) => f(ev)); if (!ev.stopped) (doc.listeners[ev.type] || []).slice().forEach((f) => f(ev)); return !ev.defaultPrevented; }
    click() { const e = new Ev('click', { cancelable: true }); this.dispatchEvent(e); }
    all(out = []) { out.push(this); this.children.forEach((c) => c.all(out)); return out; }
  }
  doc.createElement = (t) => new N(t);
  doc.documentElement = new N('html');
  doc.body = new N('body');
  doc.addEventListener = (t, f) => { (doc.listeners[t] = doc.listeners[t] || []).push(f); };
  doc.dispatchEvent = (ev) => { ev.target = ev.target || doc; (doc.listeners[ev.type] || []).slice().forEach((f) => f(ev)); return !ev.defaultPrevented; };
  doc.all = () => doc.body.all();
  doc.getElementById = (id) => doc.all().find((n) => n.id === id) || null;
  doc.querySelector = (sel) => {
    if (sel === '[data-nav-host]') return doc.all().find((n) => 'navHost' in n.dataset) || null;
    if (sel === '.d-band-status .d-links') { const b = doc.all().find((n) => (' ' + n.className + ' ').includes(' d-band-status ')); return b ? b.all().find((n) => (' ' + n.className + ' ').includes(' d-links ')) || null : null; }
    return null;
  };
  doc.N = N;
  return doc;
}
// A page like the shell pages: a band with a .d-links host, or (bare) nothing
function page(doc, kind) {
  if (kind === 'band') {
    const band = doc.createElement('header'); band.className = 'top d-band-status';
    const links = doc.createElement('div'); links.className = 'd-links'; links.dataset.navHost = '';
    const th = doc.createElement('button'); th.id = 'btn-theme'; th.textContent = 'Theme: dark';
    links.appendChild(th); band.appendChild(links); doc.body.appendChild(band);
  }
}
module.exports = { Ev, makeDoc, page };
