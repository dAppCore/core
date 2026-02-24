import { inputStyles } from './styles';
import type { LemSendDetail } from './types';

export class LemInput extends HTMLElement {
  private shadow!: ShadowRoot;
  private textarea!: HTMLTextAreaElement;
  private sendBtn!: HTMLButtonElement;
  private _disabled = false;

  constructor() {
    super();
    this.shadow = this.attachShadow({ mode: 'open' });
  }

  connectedCallback(): void {
    const style = document.createElement('style');
    style.textContent = inputStyles;

    const wrapper = document.createElement('div');
    wrapper.className = 'input-wrapper';

    this.textarea = document.createElement('textarea');
    this.textarea.rows = 1;
    this.textarea.placeholder = 'Message LEM...';

    this.sendBtn = document.createElement('button');
    this.sendBtn.className = 'send-btn';
    this.sendBtn.type = 'button';
    this.sendBtn.disabled = true;
    this.sendBtn.appendChild(this.createSendIcon());

    wrapper.appendChild(this.textarea);
    wrapper.appendChild(this.sendBtn);
    this.shadow.appendChild(style);
    this.shadow.appendChild(wrapper);

    this.textarea.addEventListener('input', () => {
      this.textarea.style.height = 'auto';
      this.textarea.style.height =
        Math.min(this.textarea.scrollHeight, 120) + 'px';
      this.sendBtn.disabled =
        this._disabled || this.textarea.value.trim() === '';
    });

    this.textarea.addEventListener('keydown', (e: KeyboardEvent) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        this.submit();
      }
    });

    this.sendBtn.addEventListener('click', () => this.submit());
  }

  private createSendIcon(): SVGSVGElement {
    const ns = 'http://www.w3.org/2000/svg';
    const svg = document.createElementNS(ns, 'svg');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('stroke', 'currentColor');
    svg.setAttribute('stroke-width', '2');
    svg.setAttribute('stroke-linecap', 'round');
    svg.setAttribute('stroke-linejoin', 'round');
    svg.setAttribute('width', '16');
    svg.setAttribute('height', '16');
    const line = document.createElementNS(ns, 'line');
    line.setAttribute('x1', '22');
    line.setAttribute('y1', '2');
    line.setAttribute('x2', '11');
    line.setAttribute('y2', '13');
    const polygon = document.createElementNS(ns, 'polygon');
    polygon.setAttribute('points', '22 2 15 22 11 13 2 9 22 2');
    svg.appendChild(line);
    svg.appendChild(polygon);
    return svg;
  }

  private submit(): void {
    const text = this.textarea.value.trim();
    if (!text || this._disabled) return;
    this.dispatchEvent(
      new CustomEvent<LemSendDetail>('lem-send', {
        bubbles: true,
        composed: true,
        detail: { text },
      })
    );
    this.textarea.value = '';
    this.textarea.style.height = 'auto';
    this.sendBtn.disabled = true;
    this.textarea.focus();
  }

  get disabled(): boolean {
    return this._disabled;
  }

  set disabled(value: boolean) {
    this._disabled = value;
    this.textarea.disabled = value;
    this.sendBtn.disabled = value || this.textarea.value.trim() === '';
    this.textarea.placeholder = value ? 'LEM is thinking...' : 'Message LEM...';
  }

  override focus(): void {
    this.textarea?.focus();
  }
}

customElements.define('lem-input', LemInput);
