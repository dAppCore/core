import { chatStyles } from './styles';
import type { ChatMessage, ChatCompletionChunk, LemSendDetail } from './types';
import { LemMessages } from './lem-messages';
import { LemInput } from './lem-input';
import './lem-message';

export class LemChat extends HTMLElement {
  private shadow!: ShadowRoot;
  private messages!: LemMessages;
  private input!: LemInput;
  private statusEl!: HTMLDivElement;
  private history: ChatMessage[] = [];
  private abortController: AbortController | null = null;

  static get observedAttributes(): string[] {
    return ['endpoint', 'model', 'system-prompt', 'max-tokens', 'temperature'];
  }

  constructor() {
    super();
    this.shadow = this.attachShadow({ mode: 'open' });
  }

  connectedCallback(): void {
    const style = document.createElement('style');
    style.textContent = chatStyles;

    const header = document.createElement('div');
    header.className = 'header';

    this.statusEl = document.createElement('div');
    this.statusEl.className = 'header-status';

    const icon = document.createElement('div');
    icon.className = 'header-icon';
    icon.textContent = 'L';

    const title = document.createElement('div');
    title.className = 'header-title';
    title.textContent = 'LEM';

    const modelLabel = document.createElement('div');
    modelLabel.className = 'header-model';
    modelLabel.textContent = this.getAttribute('model') || 'local';

    header.appendChild(this.statusEl);
    header.appendChild(icon);
    header.appendChild(title);
    header.appendChild(modelLabel);

    this.messages = document.createElement('lem-messages') as LemMessages;
    this.input = document.createElement('lem-input') as LemInput;

    this.shadow.appendChild(style);
    this.shadow.appendChild(header);
    this.shadow.appendChild(this.messages);
    this.shadow.appendChild(this.input);

    this.addEventListener('lem-send', ((e: Event) => {
      const detail = (e as CustomEvent<LemSendDetail>).detail;
      this.handleSend(detail.text);
    }) as EventListener);

    const systemPrompt = this.getAttribute('system-prompt');
    if (systemPrompt) {
      this.history.push({ role: 'system', content: systemPrompt });
    }

    this.checkConnection();
    requestAnimationFrame(() => this.input.focus());
  }

  disconnectedCallback(): void {
    this.abortController?.abort();
  }

  get endpoint(): string {
    const attr = this.getAttribute('endpoint');
    if (!attr) return window.location.origin;
    return attr;
  }

  get model(): string {
    return this.getAttribute('model') || '';
  }

  get maxTokens(): number {
    const val = this.getAttribute('max-tokens');
    return val ? parseInt(val, 10) : 2048;
  }

  get temperature(): number {
    const val = this.getAttribute('temperature');
    return val ? parseFloat(val) : 0.7;
  }

  private async checkConnection(): Promise<void> {
    try {
      const resp = await fetch(`${this.endpoint}/v1/models`, {
        signal: AbortSignal.timeout(3000),
      });
      this.statusEl.classList.toggle('disconnected', !resp.ok);
    } catch {
      this.statusEl.classList.add('disconnected');
    }
  }

  private async handleSend(text: string): Promise<void> {
    this.messages.addMessage('user', text);
    this.history.push({ role: 'user', content: text });

    const assistantMsg = this.messages.addMessage('assistant');
    assistantMsg.streaming = true;
    this.input.disabled = true;

    this.abortController?.abort();
    this.abortController = new AbortController();

    let fullResponse = '';

    try {
      const response = await fetch(`${this.endpoint}/v1/chat/completions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        signal: this.abortController.signal,
        body: JSON.stringify({
          model: this.model,
          messages: this.history,
          max_tokens: this.maxTokens,
          temperature: this.temperature,
          stream: true,
        }),
      });

      if (!response.ok) {
        throw new Error(`Server error: ${response.status}`);
      }
      if (!response.body) {
        throw new Error('No response body');
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue;
          const data = line.slice(6).trim();
          if (data === '[DONE]') continue;

          try {
            const chunk: ChatCompletionChunk = JSON.parse(data);
            const delta = chunk.choices?.[0]?.delta;
            if (delta?.content) {
              fullResponse += delta.content;
              assistantMsg.appendToken(delta.content);
              this.messages.scrollToBottom();
            }
          } catch {
            // skip malformed chunks
          }
        }
      }
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        // user-initiated abort — ignore
      } else {
        const errorText =
          err instanceof Error ? err.message : 'Connection failed';
        if (!fullResponse) {
          assistantMsg.text = `\u26A0\uFE0F ${errorText}`;
        }
        this.statusEl.classList.add('disconnected');
      }
    } finally {
      assistantMsg.streaming = false;
      this.input.disabled = false;
      this.input.focus();
      this.abortController = null;
      if (fullResponse) {
        this.history.push({ role: 'assistant', content: fullResponse });
      }
    }
  }
}

customElements.define('lem-chat', LemChat);
