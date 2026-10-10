export function el(tag, attrs = {}, ...children) {
  const node = document.createElement(tag);
  for (const [key, value] of Object.entries(attrs)) {
    if (value === null || value === undefined || value === false) continue;
    if (key === 'className') node.className = value;
    else if (key.startsWith('on') && typeof value === 'function') node.addEventListener(key.slice(2).toLowerCase(), value);
    else if (key === 'value') node.value = value;
    else if (key === 'checked') node.checked = value;
    else node.setAttribute(key, value === true ? '' : String(value));
  }
  for (const child of children.flat(Infinity)) {
    if (child === null || child === undefined || child === false) continue;
    node.append(child instanceof Node ? child : document.createTextNode(String(child)));
  }
  return node;
}
export const button = (text, action, attrs = {}) => el('button', { type: 'button', onClick: action, ...attrs }, text);
export const link = (text, route, params = {}) => el('a', { href: `#${route}${Object.keys(params).length ? '?' + new URLSearchParams(params) : ''}` }, text);
export const heading = (title, subtitle) => el('div', { className: 'page-heading' }, el('h1', {}, title), subtitle ? el('p', { className: 'lede' }, subtitle) : null);
export const notice = (text, tone = 'info') => el('p', { className: `notice ${tone}`, role: tone === 'error' ? 'alert' : 'status' }, text);
export const empty = (title, detail, action) => el('section', { className: 'empty-state' }, el('h2', {}, title), el('p', {}, detail), action);
export function field(label, control, hint) {
  const wrapper = el('label', { className: 'field' }, el('span', {}, label), control);
  if (hint) wrapper.append(el('small', {}, hint));
  return wrapper;
}
export function table(headers, rows, caption) {
  return el('div', { className: 'table-wrap' }, el('table', {}, caption ? el('caption', {}, caption) : null,
    el('thead', {}, el('tr', {}, headers.map(h => el('th', { scope: 'col' }, h)))),
    el('tbody', {}, rows.map(row => el('tr', {}, row.map((value, index) => el(index === 0 ? 'th' : 'td', index === 0 ? { scope: 'row' } : {}, value)))))));
}
export function value(input, fallback = 'Unknown') {
  return input === null || input === undefined || (typeof input === 'number' && !Number.isFinite(input)) ? fallback : String(input);
}
export function source(result) {
  return el('details', { className: 'source-details' }, el('summary', {}, 'Evidence source'),
    el('dl', {}, el('dt', {}, 'Source'), el('dd', {}, result?.source || 'Unknown'),
      el('dt', {}, 'Retrieved by MCP'), el('dd', {}, result?.retrievedAt || 'Unknown'),
      el('dt', {}, 'Limits'), el('dd', {}, result?.note || 'Capture time must be read from source data; retrieval time is not capture time.')));
}
