import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { WailsService, ConnectionStatus } from '@shared/wails.service';

@Component({
  selector: 'app-tray',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="tray">
      <div class="tray__header">
        <h3>Core IDE</h3>
        <span class="badge" [class]="status.bridgeConnected ? 'badge--success' : 'badge--danger'">
          {{ status.bridgeConnected ? 'Online' : 'Offline' }}
        </span>
      </div>

      <div class="tray__stats">
        <div class="stat">
          <span class="stat__value">{{ status.wsClients }}</span>
          <span class="stat__label">WS Clients</span>
        </div>
        <div class="stat">
          <span class="stat__value">{{ status.wsChannels }}</span>
          <span class="stat__label">Channels</span>
        </div>
      </div>

      <div class="tray__actions">
        <button class="btn btn--primary" (click)="openMain()">Open IDE</button>
        <button class="btn btn--secondary" (click)="openSettings()">Settings</button>
      </div>

      <div class="tray__footer text-muted">
        Laravel bridge: {{ status.bridgeConnected ? 'connected' : 'disconnected' }}
      </div>
    </div>
  `,
  styles: [`
    .tray {
      padding: var(--spacing-md);
      height: 100%;
      display: flex;
      flex-direction: column;
      gap: var(--spacing-md);
    }

    .tray__header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .tray__stats {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: var(--spacing-sm);
    }

    .stat {
      background: var(--bg-secondary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      padding: var(--spacing-sm) var(--spacing-md);
      text-align: center;
    }

    .stat__value {
      display: block;
      font-size: 24px;
      font-weight: 600;
      color: var(--accent-primary);
    }

    .stat__label {
      font-size: 12px;
      color: var(--text-muted);
    }

    .tray__actions {
      display: flex;
      gap: var(--spacing-sm);
    }

    .tray__actions .btn {
      flex: 1;
    }

    .tray__footer {
      margin-top: auto;
      font-size: 12px;
      text-align: center;
    }
  `]
})
export class TrayComponent implements OnInit {
  status: ConnectionStatus = {
    bridgeConnected: false,
    laravelUrl: '',
    wsClients: 0,
    wsChannels: 0
  };

  private pollTimer: ReturnType<typeof setInterval> | null = null;

  constructor(private wails: WailsService) {}

  ngOnInit(): void {
    this.refresh();
    this.pollTimer = setInterval(() => this.refresh(), 5000);
  }

  async refresh(): Promise<void> {
    this.status = await this.wails.getConnectionStatus();
  }

  openMain(): void {
    this.wails.showWindow('main');
  }

  openSettings(): void {
    this.wails.showWindow('settings');
  }
}
