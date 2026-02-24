import { messageStyles } from './styles';
import { renderMarkdown } from './markdown';

interface ThinkSplit {
  think: string | null;
  response: string;
}

export class LemMessage extends HTMLElement {
  private shadow!: ShadowRoot;
  private thinkPanel!: HTMLDivElement;
  private thinkContent!: HTMLDivElement;
  private thinkLabel!: HTMLDivElement;
  private contentEl!: HTMLDivElement;
  private cursorEl: HTMLSpanElement | null = null;
  private _text = '';
  private _streaming = false;
  private _thinkCollapsed = false;

  constructor() {
    super();
    this.shadow = this.attachShadow({ mode: 'open' });
  }

  connectedCallback(): void {
    const role = this.getAttribute('role') || 'user';

    const style = document.createElement('style');
    style.textContent = messageStyles;

    const bubble = document.createElement('div');
    bubble.className = 'bubble';

    const roleEl = document.createElement('div');
    roleEl.className = 'role';
    roleEl.textContent = role === 'assistant' ? 'LEM' : 'You';

    this.thinkPanel = document.createElement('div');
    this.thinkPanel.className = 'think-panel';
    this.thinkPanel.style.display = 'none';

    this.thinkLabel = document.createElement('div');
    this.thinkLabel.className = 'think-label';
    this.thinkLabel.textContent = '\u25BC reasoning';
    this.thinkLabel.addEventListener('click', () => {
      this._thinkCollapsed = !this._thinkCollapsed;
      this.thinkPanel.classList.toggle('collapsed', this._thinkCollapsed);
      this.thinkLabel.textContent = this._thinkCollapsed
        ? '\u25B6 reasoning'
        : '\u25BC reasoning';
    });

    this.thinkContent = document.createElement('div');
    this.thinkContent.className = 'think-content';
    this.thinkPanel.appendChild(this.thinkLabel);
    this.thinkPanel.appendChild(this.thinkContent);

    this.contentEl = document.createElement('div');
    this.contentEl.className = 'content';

    bubble.appendChild(roleEl);
    if (role === 'assistant') {
      bubble.appendChild(this.thinkPanel);
    }
    bubble.appendChild(this.contentEl);

    this.shadow.appendChild(style);
    this.shadow.appendChild(bubble);

    if (this._text) {
      this.updateContent();
    }
  }

  get text(): string {
    return this._text;
  }

  set text(value: string) {
    this._text = value;
    this.updateContent();
  }

  get streaming(): boolean {
    return this._streaming;
  }

  set streaming(value: boolean) {
    this._streaming = value;
    this.updateContent();
  }

  appendToken(token: string): void {
    this._text += token;
    this.updateContent();
  }

  /**
   * Splits text into think/response portions and renders each.
   *
   * Safety: renderMarkdown() escapes all HTML entities (& < > ") before any
   * inline formatting is applied. The source is the local MLX model output,
   * not arbitrary user HTML. Shadow DOM provides additional isolation.
   */
  private updateContent(): void {
    if (!this.contentEl) return;
    const { think, response } = this.splitThink(this._text);

    if (think !== null && this.thinkPanel) {
      this.thinkPanel.style.display = '';
      this.thinkContent.textContent = think;
    }

    // renderMarkdown() escapes all HTML before formatting — safe for innerHTML
    // within Shadow DOM isolation, sourced from local MLX model only
    const responseHtml = renderMarkdown(response);
    this.contentEl.innerHTML = responseHtml;

    if (this._streaming) {
      if (!this.cursorEl) {
        this.cursorEl = document.createElement('span');
        this.cursorEl.className = 'cursor';
      }
      if (think !== null && !this._text.includes('</think>')) {
        this.thinkContent.appendChild(this.cursorEl);
      } else {
        const lastChild = this.contentEl.lastElementChild || this.contentEl;
        lastChild.appendChild(this.cursorEl);
      }
    }
  }

  private splitThink(text: string): ThinkSplit {
    const thinkStart = text.indexOf('<think>');
    if (thinkStart === -1) {
      return { think: null, response: text };
    }
    const afterOpen = thinkStart + '<think>'.length;
    const thinkEnd = text.indexOf('</think>', afterOpen);
    if (thinkEnd === -1) {
      return {
        think: text.slice(afterOpen).trim(),
        response: text.slice(0, thinkStart).trim(),
      };
    }
    const thinkText = text.slice(afterOpen, thinkEnd).trim();
    const beforeThink = text.slice(0, thinkStart).trim();
    const afterThink = text.slice(thinkEnd + '</think>'.length).trim();
    const response = [beforeThink, afterThink].filter(Boolean).join('\n');
    return { think: thinkText, response };
  }
}

customElements.define('lem-message', LemMessage);
