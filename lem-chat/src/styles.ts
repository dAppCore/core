export const chatStyles = `
  :host {
    display: flex;
    flex-direction: column;
    background: var(--lem-bg, #1a1a1e);
    color: var(--lem-text, #e0e0e0);
    font-family: var(--lem-font, system-ui, -apple-system, sans-serif);
    font-size: 14px;
    line-height: 1.5;
    border-radius: 12px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  .header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 14px 18px;
    background: rgba(255, 255, 255, 0.03);
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
    flex-shrink: 0;
  }

  .header-icon {
    width: 28px;
    height: 28px;
    border-radius: 8px;
    background: var(--lem-accent, #5865f2);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 700;
    color: #fff;
  }

  .header-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--lem-text, #e0e0e0);
  }

  .header-model {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.35);
    margin-left: auto;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  }

  .header-status {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #43b581;
    flex-shrink: 0;
  }

  .header-status.disconnected {
    background: #f04747;
  }
`;

export const messagesStyles = `
  :host {
    display: block;
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 16px 0;
    scroll-behavior: smooth;
  }

  :host::-webkit-scrollbar {
    width: 6px;
  }

  :host::-webkit-scrollbar-track {
    background: transparent;
  }

  :host::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.12);
    border-radius: 3px;
  }

  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 12px;
    color: rgba(255, 255, 255, 0.25);
  }

  .empty-icon {
    font-size: 36px;
    opacity: 0.4;
  }

  .empty-text {
    font-size: 14px;
  }
`;

export const messageStyles = `
  :host {
    display: block;
    padding: 6px 18px;
  }

  :host([role="user"]) .bubble {
    background: var(--lem-msg-user, #2a2a3e);
    margin-left: 40px;
    border-radius: 12px 12px 4px 12px;
  }

  :host([role="assistant"]) .bubble {
    background: var(--lem-msg-assistant, #1e1e2a);
    margin-right: 40px;
    border-radius: 12px 12px 12px 4px;
  }

  .bubble {
    padding: 10px 14px;
    word-wrap: break-word;
    overflow-wrap: break-word;
  }

  .role {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 4px;
    color: rgba(255, 255, 255, 0.35);
  }

  :host([role="assistant"]) .role {
    color: var(--lem-accent, #5865f2);
  }

  .content {
    color: var(--lem-text, #e0e0e0);
    line-height: 1.6;
  }

  .content p {
    margin: 0 0 8px 0;
  }

  .content p:last-child {
    margin-bottom: 0;
  }

  .content strong {
    font-weight: 600;
    color: #fff;
  }

  .content em {
    font-style: italic;
    color: rgba(255, 255, 255, 0.8);
  }

  .content code {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 12px;
    background: rgba(0, 0, 0, 0.3);
    padding: 2px 5px;
    border-radius: 4px;
    color: #e8a0bf;
  }

  .content pre {
    margin: 8px 0;
    padding: 12px;
    background: rgba(0, 0, 0, 0.35);
    border-radius: 8px;
    overflow-x: auto;
    border: 1px solid rgba(255, 255, 255, 0.06);
  }

  .content pre code {
    background: none;
    padding: 0;
    font-size: 12px;
    color: #c9d1d9;
    line-height: 1.5;
  }

  .think-panel {
    margin: 6px 0 8px;
    padding: 8px 12px;
    background: rgba(88, 101, 242, 0.06);
    border-left: 2px solid rgba(88, 101, 242, 0.3);
    border-radius: 0 6px 6px 0;
    font-size: 12px;
    color: rgba(255, 255, 255, 0.45);
    line-height: 1.5;
    max-height: 200px;
    overflow-y: auto;
  }

  .think-panel::-webkit-scrollbar {
    width: 4px;
  }

  .think-panel::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 2px;
  }

  .think-label {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: rgba(88, 101, 242, 0.5);
    margin-bottom: 4px;
    cursor: pointer;
    user-select: none;
  }

  .think-label:hover {
    color: rgba(88, 101, 242, 0.7);
  }

  .think-panel.collapsed .think-content {
    display: none;
  }

  .cursor {
    display: inline-block;
    width: 7px;
    height: 16px;
    background: var(--lem-accent, #5865f2);
    border-radius: 1px;
    animation: blink 0.8s step-end infinite;
    vertical-align: text-bottom;
    margin-left: 2px;
  }

  @keyframes blink {
    50% { opacity: 0; }
  }
`;

export const inputStyles = `
  :host {
    display: block;
    padding: 12px 16px 16px;
    border-top: 1px solid rgba(255, 255, 255, 0.06);
    flex-shrink: 0;
  }

  .input-wrapper {
    display: flex;
    align-items: flex-end;
    gap: 10px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    padding: 8px 12px;
    transition: border-color 0.15s;
  }

  .input-wrapper:focus-within {
    border-color: var(--lem-accent, #5865f2);
  }

  textarea {
    flex: 1;
    background: none;
    border: none;
    outline: none;
    color: var(--lem-text, #e0e0e0);
    font-family: inherit;
    font-size: 14px;
    line-height: 1.5;
    resize: none;
    max-height: 120px;
    min-height: 22px;
    padding: 0;
  }

  textarea::placeholder {
    color: rgba(255, 255, 255, 0.25);
  }

  .send-btn {
    background: var(--lem-accent, #5865f2);
    border: none;
    border-radius: 8px;
    color: #fff;
    width: 32px;
    height: 32px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: opacity 0.15s, transform 0.1s;
  }

  .send-btn:hover {
    opacity: 0.85;
  }

  .send-btn:active {
    transform: scale(0.95);
  }

  .send-btn:disabled {
    opacity: 0.3;
    cursor: default;
    transform: none;
  }

  .send-btn svg {
    width: 16px;
    height: 16px;
  }
`;
