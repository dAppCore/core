import { messagesStyles } from './styles';
import type { LemMessage } from './lem-message';

export class LemMessages extends HTMLElement {
  private shadow!: ShadowRoot;
  private container!: HTMLDivElement;
  private emptyEl!: HTMLDivElement;
  private shouldAutoScroll = true;

  constructor() {
    super();
    this.shadow = this.attachShadow({ mode: 'open' });
  }

  connectedCallback(): void {
    const style = document.createElement('style');
    style.textContent = messagesStyles;

    this.container = document.createElement('div');

    this.emptyEl = document.createElement('div');
    this.emptyEl.className = 'empty';
    const emptyIcon = document.createElement('div');
    emptyIcon.className = 'empty-icon';
    emptyIcon.textContent = '\u2728';
    const emptyText = document.createElement('div');
    emptyText.className = 'empty-text';
    emptyText.textContent = 'Start a conversation';
    this.emptyEl.appendChild(emptyIcon);
    this.emptyEl.appendChild(emptyText);

    this.shadow.appendChild(style);
    this.shadow.appendChild(this.emptyEl);
    this.shadow.appendChild(this.container);

    this.addEventListener('scroll', () => {
      const threshold = 60;
      this.shouldAutoScroll =
        this.scrollHeight - this.scrollTop - this.clientHeight < threshold;
    });
  }

  addMessage(role: string, text?: string): LemMessage {
    this.emptyEl.style.display = 'none';
    const msg = document.createElement('lem-message') as LemMessage;
    msg.setAttribute('role', role);
    this.container.appendChild(msg);
    if (text) {
      msg.text = text;
    }
    this.scrollToBottom();
    return msg;
  }

  scrollToBottom(): void {
    if (this.shouldAutoScroll) {
      requestAnimationFrame(() => {
        this.scrollTop = this.scrollHeight;
      });
    }
  }

  clear(): void {
    this.container.replaceChildren();
    this.emptyEl.style.display = '';
    this.shouldAutoScroll = true;
  }
}

customElements.define('lem-messages', LemMessages);
