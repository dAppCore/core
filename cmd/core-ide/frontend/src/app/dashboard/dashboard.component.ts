import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { WailsService, DashboardData } from '@shared/wails.service';
import { WebSocketService, WSMessage } from '@shared/ws.service';
import { Subscription } from 'rxjs';

interface ActivityItem {
  type: string;
  message: string;
  timestamp: string;
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="dashboard">
      <h2>Dashboard</h2>

      <div class="dashboard__grid">
        <div class="stat-card">
          <div class="stat-card__value" [class.text-success]="data.connection.bridgeConnected">
            {{ data.connection.bridgeConnected ? 'Online' : 'Offline' }}
          </div>
          <div class="stat-card__label">Bridge Status</div>
        </div>
        <div class="stat-card">
          <div class="stat-card__value">{{ data.connection.wsClients }}</div>
          <div class="stat-card__label">WS Clients</div>
        </div>
        <div class="stat-card">
          <div class="stat-card__value">{{ data.connection.wsChannels }}</div>
          <div class="stat-card__label">Active Channels</div>
        </div>
        <div class="stat-card">
          <div class="stat-card__value">0</div>
          <div class="stat-card__label">Agent Sessions</div>
        </div>
      </div>

      <div class="dashboard__activity">
        <h3>Activity Feed</h3>
        <div class="activity-feed">
          <div *ngFor="let item of activity" class="activity-item">
            <span class="activity-item__badge badge badge--info">{{ item.type }}</span>
            <span class="activity-item__msg">{{ item.message }}</span>
            <span class="activity-item__time text-muted">{{ item.timestamp | date:'shortTime' }}</span>
          </div>
          <div *ngIf="activity.length === 0" class="text-muted" style="text-align: center; padding: var(--spacing-lg);">
            No recent activity. Events will stream here in real-time.
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .dashboard {
      padding: var(--spacing-md);
    }

    .dashboard__grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
      gap: var(--spacing-md);
      margin: var(--spacing-md) 0;
    }

    .stat-card {
      background: var(--bg-secondary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: var(--spacing-lg);
      text-align: center;
    }

    .stat-card__value {
      font-size: 28px;
      font-weight: 700;
      color: var(--accent-primary);
    }

    .stat-card__label {
      font-size: 13px;
      color: var(--text-muted);
      margin-top: var(--spacing-xs);
    }

    .dashboard__activity {
      margin-top: var(--spacing-lg);
    }

    .activity-feed {
      margin-top: var(--spacing-sm);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      max-height: 400px;
      overflow-y: auto;
    }

    .activity-item {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);
      padding: var(--spacing-sm) var(--spacing-md);
      border-bottom: 1px solid var(--border-color);
      font-size: 13px;

      &:last-child {
        border-bottom: none;
      }
    }

    .activity-item__msg {
      flex: 1;
    }

    .activity-item__time {
      font-size: 12px;
      white-space: nowrap;
    }
  `]
})
export class DashboardComponent implements OnInit, OnDestroy {
  data: DashboardData = {
    connection: { bridgeConnected: false, laravelUrl: '', wsClients: 0, wsChannels: 0 }
  };
  activity: ActivityItem[] = [];

  private sub: Subscription | null = null;
  private pollTimer: ReturnType<typeof setInterval> | null = null;

  constructor(
    private wails: WailsService,
    private wsService: WebSocketService
  ) {}

  ngOnInit(): void {
    this.refresh();
    this.pollTimer = setInterval(() => this.refresh(), 10000);

    this.wsService.connect();
    this.sub = this.wsService.subscribe('dashboard:activity').subscribe(
      (msg: WSMessage) => {
        if (msg.data && typeof msg.data === 'object') {
          this.activity.unshift(msg.data as ActivityItem);
          if (this.activity.length > 100) {
            this.activity.pop();
          }
        }
      }
    );
  }

  ngOnDestroy(): void {
    this.sub?.unsubscribe();
    if (this.pollTimer) clearInterval(this.pollTimer);
  }

  async refresh(): Promise<void> {
    this.data = await this.wails.getDashboard();
  }
}
