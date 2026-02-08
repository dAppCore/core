import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DomSanitizer, SafeResourceUrl } from '@angular/platform-browser';

type Mode = 'web' | 'stream';

@Component({
  selector: 'app-jellyfin',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="jellyfin">
      <header class="jellyfin__header">
        <div>
          <h1>Jellyfin Player</h1>
          <p class="text-muted">Quick embed for media.lthn.ai or any Jellyfin host.</p>
        </div>
        <div class="mode-switch">
          <button class="btn btn--secondary" [class.is-active]="mode === 'web'" (click)="mode = 'web'">Web</button>
          <button class="btn btn--secondary" [class.is-active]="mode === 'stream'" (click)="mode = 'stream'">Stream</button>
        </div>
      </header>

      <div class="card jellyfin__config">
        <div class="form-group">
          <label class="form-label">Jellyfin Server URL</label>
          <input class="form-input" [(ngModel)]="serverUrl" placeholder="https://media.lthn.ai" />
        </div>

        <div *ngIf="mode === 'stream'" class="stream-grid">
          <div class="form-group">
            <label class="form-label">Item ID</label>
            <input class="form-input" [(ngModel)]="itemId" placeholder="Jellyfin library item ID" />
          </div>
          <div class="form-group">
            <label class="form-label">API Key</label>
            <input class="form-input" [(ngModel)]="apiKey" placeholder="Jellyfin API key" />
          </div>
          <div class="form-group">
            <label class="form-label">Media Source ID (optional)</label>
            <input class="form-input" [(ngModel)]="mediaSourceId" placeholder="Source ID for multi-source items" />
          </div>
        </div>

        <div class="actions">
          <button class="btn btn--primary" (click)="load()">Load Player</button>
          <button class="btn btn--secondary" (click)="reset()">Reset</button>
        </div>
      </div>

      <div class="card jellyfin__viewer" *ngIf="loaded && mode === 'web'">
        <iframe
          class="jellyfin-frame"
          title="Jellyfin Web"
          [src]="safeWebUrl"
          loading="lazy"
          referrerpolicy="no-referrer"
        ></iframe>
      </div>

      <div class="card jellyfin__viewer" *ngIf="loaded && mode === 'stream'">
        <video class="jellyfin-video" controls [src]="streamUrl"></video>
        <p class="text-muted stream-hint" *ngIf="!streamUrl">Set Item ID and API key to build stream URL.</p>
      </div>
    </div>
  `,
  styles: [`
    .jellyfin {
      display: flex;
      flex-direction: column;
      gap: var(--spacing-md);
      padding: var(--spacing-md);
      height: 100%;
      overflow: auto;
      background: var(--bg-secondary);
    }

    .jellyfin__header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: var(--spacing-md);
    }

    .jellyfin__header h1 {
      margin-bottom: var(--spacing-xs);
    }

    .mode-switch {
      display: flex;
      gap: var(--spacing-xs);
    }

    .mode-switch .btn.is-active {
      border-color: var(--accent-primary);
      color: var(--accent-primary);
    }

    .jellyfin__config {
      display: flex;
      flex-direction: column;
      gap: var(--spacing-sm);
    }

    .stream-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
      gap: var(--spacing-sm);
    }

    .actions {
      display: flex;
      gap: var(--spacing-sm);
    }

    .jellyfin__viewer {
      flex: 1;
      min-height: 420px;
      padding: 0;
      overflow: hidden;
    }

    .jellyfin-frame,
    .jellyfin-video {
      border: 0;
      width: 100%;
      height: 100%;
      min-height: 420px;
      background: #000;
    }

    .stream-hint {
      padding: var(--spacing-md);
      margin: 0;
    }
  `]
})
export class JellyfinComponent {
  mode: Mode = 'web';
  loaded = false;

  serverUrl = 'https://media.lthn.ai';
  itemId = '';
  apiKey = '';
  mediaSourceId = '';

  safeWebUrl!: SafeResourceUrl;
  streamUrl = '';

  constructor(private sanitizer: DomSanitizer) {
    this.safeWebUrl = this.sanitizer.bypassSecurityTrustResourceUrl('https://media.lthn.ai/web/index.html');
  }

  load(): void {
    const base = this.normalizeBase(this.serverUrl);
    this.safeWebUrl = this.sanitizer.bypassSecurityTrustResourceUrl(`${base}/web/index.html`);
    this.streamUrl = this.buildStreamUrl(base);
    this.loaded = true;
  }

  reset(): void {
    this.loaded = false;
    this.itemId = '';
    this.apiKey = '';
    this.mediaSourceId = '';
    this.streamUrl = '';
  }

  private normalizeBase(value: string): string {
    const raw = value.trim() || 'https://media.lthn.ai';
    const withProtocol = raw.startsWith('http://') || raw.startsWith('https://') ? raw : `https://${raw}`;
    return withProtocol.replace(/\/+$/, '');
  }

  private buildStreamUrl(base: string): string {
    if (!this.itemId.trim() || !this.apiKey.trim()) {
      return '';
    }

    const url = new URL(`${base}/Videos/${encodeURIComponent(this.itemId.trim())}/stream`);
    url.searchParams.set('api_key', this.apiKey.trim());
    url.searchParams.set('static', 'true');
    if (this.mediaSourceId.trim()) {
      url.searchParams.set('MediaSourceId', this.mediaSourceId.trim());
    }
    return url.toString();
  }
}
