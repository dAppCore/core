function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function parseInline(text: string): string {
  let result = escapeHtml(text);
  result = result.replace(/`([^`]+)`/g, '<code>$1</code>');
  result = result.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  result = result.replace(/__(.+?)__/g, '<strong>$1</strong>');
  result = result.replace(/(?<!\w)\*([^*]+)\*(?!\w)/g, '<em>$1</em>');
  result = result.replace(/(?<!\w)_([^_]+)_(?!\w)/g, '<em>$1</em>');
  return result;
}

function wrapParagraph(lines: string[]): string {
  const joined = lines.join('<br>');
  if (joined.startsWith('<pre')) return joined;
  return `<p>${joined}</p>`;
}

export function renderMarkdown(text: string): string {
  const lines = text.split('\n');
  const output: string[] = [];
  let inCodeBlock = false;
  let codeLines: string[] = [];
  let codeLang = '';

  for (const line of lines) {
    if (line.trimStart().startsWith('```')) {
      if (!inCodeBlock) {
        inCodeBlock = true;
        codeLang = line.trimStart().slice(3).trim();
        codeLines = [];
      } else {
        const langAttr = codeLang ? ` data-lang="${escapeHtml(codeLang)}"` : '';
        output.push(`<pre${langAttr}><code>${escapeHtml(codeLines.join('\n'))}</code></pre>`);
        inCodeBlock = false;
        codeLines = [];
        codeLang = '';
      }
      continue;
    }
    if (inCodeBlock) {
      codeLines.push(line);
      continue;
    }
    if (line.trim() === '') {
      output.push('');
      continue;
    }
    output.push(parseInline(line));
  }

  if (inCodeBlock) {
    const langAttr = codeLang ? ` data-lang="${escapeHtml(codeLang)}"` : '';
    output.push(`<pre${langAttr}><code>${escapeHtml(codeLines.join('\n'))}</code></pre>`);
  }

  const paragraphs: string[] = [];
  let current: string[] = [];
  for (const line of output) {
    if (line === '') {
      if (current.length > 0) {
        paragraphs.push(wrapParagraph(current));
        current = [];
      }
    } else {
      current.push(line);
    }
  }
  if (current.length > 0) {
    paragraphs.push(wrapParagraph(current));
  }

  return paragraphs.join('');
}
