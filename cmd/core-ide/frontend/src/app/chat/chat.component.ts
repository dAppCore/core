import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { WailsService, ChatMessage, Session, PlanStatus } from '@shared/wails.service';
import { WebSocketService, WSMessage } from '@shared/ws.service';
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-chat',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="chat">
      <div class="chat__header">
        <div class="chat__session-picker">
          <select class="form-select" [(ngModel)]="activeSessionId" (ngModelChange)="onSessionChange()">
            <option *ngFor="let s of sessions" [value]="s.id">{{ s.name }} ({{ s.status }})</option>
          </select>
          <button class="btn btn--ghost" (click)="createSession()">+ New</button>
        </div>
      </div>

      <div class="chat__body">
        <div class="chat__messages">
          <div
            *ngFor="let msg of messages"
            class="chat__msg"
            [class.chat__msg--user]="msg.role === 'user'"
            [class.chat__msg--agent]="msg.role === 'agent'"
          >
            <div class="chat__msg-role">{{ msg.role }}</div>
            <div class="chat__msg-content">{{ msg.content }}</div>
          </div>
          <div *ngIf="messages.length === 0" class="chat__empty text-muted">
            No messages yet. Start a conversation with an agent.
          </div>
        </div>

        <div *ngIf="plan.steps.length > 0" class="chat__plan">
          <h4>Plan: {{ plan.status }}</h4>
          <ul>
            <li *ngFor="let step of plan.steps" [class]="'plan-step plan-step--' + step.status">
              {{ step.name }}
              <span class="badge badge--info">{{ step.status }}</span>
            </li>
          </ul>
        </div>
      </div>

      <div class="chat__input">
        <textarea
          class="form-textarea"
          [(ngModel)]="draft"
          (keydown.enter)="sendMessage($event)"
          placeholder="Type a message... (Enter to send)"
          rows="2"
        ></textarea>
        <button class="btn btn--primary" (click)="sendMessage()" [disabled]="!draft.trim()">Send</button>
      </div>
    </div>
  `,
  styles: [`
    .chat {
      display: flex;
      flex-direction: column;
      height: 100%;
    }

    .chat__header {
      padding: var(--spacing-sm) var(--spacing-md);
      border-bottom: 1px solid var(--border-color);
    }

    .chat__session-picker {
      display: flex;
      gap: var(--spacing-sm);
      align-items: center;
    }

    .chat__session-picker select {
      flex: 1;
    }

    .chat__body {
      flex: 1;
      display: flex;
      overflow: hidden;
    }

    .chat__messages {
      flex: 1;
      overflow-y: auto;
      padding: var(--spacing-md);
      display: flex;
      flex-direction: column;
      gap: var(--spacing-sm);
    }

    .chat__msg {
      padding: var(--spacing-sm) var(--spacing-md);
      border-radius: var(--radius-md);
      max-width: 80%;
    }

    .chat__msg--user {
      align-self: flex-end;
      background: rgba(57, 208, 216, 0.12);
      border: 1px solid rgba(57, 208, 216, 0.2);
    }

    .chat__msg--agent {
      align-self: flex-start;
      background: var(--bg-secondary);
      border: 1px solid var(--border-color);
    }

    .chat__msg-role {
      font-size: 11px;
      font-weight: 600;
      text-transform: uppercase;
      color: var(--text-muted);
      margin-bottom: 2px;
    }

    .chat__msg-content {
      white-space: pre-wrap;
      word-break: break-word;
    }

    .chat__empty {
      margin: auto;
      text-align: center;
    }

    .chat__plan {
      width: 260px;
      border-left: 1px solid var(--border-color);
      padding: var(--spacing-md);
      overflow-y: auto;
    }

    .chat__plan ul {
      list-style: none;
      margin-top: var(--spacing-sm);
    }

    .chat__plan li {
      padding: var(--spacing-xs) 0;
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-size: 13px;
    }

    .chat__input {
      padding: var(--spacing-sm) var(--spacing-md);
      border-top: 1px solid var(--border-color);
      display: flex;
      gap: var(--spacing-sm);
      align-items: flex-end;
    }

    .chat__input textarea {
      flex: 1;
      resize: none;
    }
  `]
})
export class ChatComponent implements OnInit, OnDestroy {
  sessions: Session[] = [];
  activeSessionId = '';
  messages: ChatMessage[] = [];
  plan: PlanStatus = { sessionId: '', status: '', steps: [] };
  draft = '';

  private sub: Subscription | null = null;

  constructor(
    private wails: WailsService,
    private wsService: WebSocketService
  ) {}

  ngOnInit(): void {
    this.loadSessions();
    this.wsService.connect();
  }

  ngOnDestroy(): void {
    this.sub?.unsubscribe();
  }

  async loadSessions(): Promise<void> {
    this.sessions = await this.wails.listSessions();
    if (this.sessions.length > 0 && !this.activeSessionId) {
      this.activeSessionId = this.sessions[0].id;
      this.onSessionChange();
    }
  }

  async onSessionChange(): Promise<void> {
    if (!this.activeSessionId) return;

    // Unsubscribe from previous channel
    this.sub?.unsubscribe();

    // Load history and plan
    this.messages = await this.wails.getHistory(this.activeSessionId);
    this.plan = await this.wails.getPlanStatus(this.activeSessionId);

    // Subscribe to live updates
    this.sub = this.wsService.subscribe(`chat:${this.activeSessionId}`).subscribe(
      (msg: WSMessage) => {
        if (msg.data && typeof msg.data === 'object') {
          this.messages.push(msg.data as ChatMessage);
        }
      }
    );
  }

  async sendMessage(event?: KeyboardEvent): Promise<void> {
    if (event) {
      if (event.shiftKey) return; // Allow shift+enter for newlines
      event.preventDefault();
    }
    const text = this.draft.trim();
    if (!text || !this.activeSessionId) return;

    // Optimistic UI update
    this.messages.push({ role: 'user', content: text, timestamp: new Date().toISOString() });
    this.draft = '';

    await this.wails.sendMessage(this.activeSessionId, text);
  }

  async createSession(): Promise<void> {
    const name = `Session ${this.sessions.length + 1}`;
    const session = await this.wails.createSession(name);
    this.sessions.push(session);
    this.activeSessionId = session.id;
    this.onSessionChange();
  }
}
