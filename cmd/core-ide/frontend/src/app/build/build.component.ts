import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { WailsService, Build } from '@shared/wails.service';
import { WebSocketService, WSMessage } from '@shared/ws.service';
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-build',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="builds">
      <div class="builds__header">
        <h2>Builds</h2>
        <button class="btn btn--secondary" (click)="refresh()">Refresh</button>
      </div>

      <div class="builds__list">
        <div
          *ngFor="let build of builds; trackBy: trackBuild"
          class="build-card"
          [class.build-card--expanded]="expandedId === build.id"
          (click)="toggle(build.id)"
        >
          <div class="build-card__header">
            <div class="build-card__info">
              <span class="build-card__repo">{{ build.repo }}</span>
              <span class="build-card__branch text-muted">{{ build.branch }}</span>
            </div>
            <span class="badge" [class]="statusBadge(build.status)">{{ build.status }}</span>
          </div>

          <div class="build-card__meta text-muted">
            {{ build.startedAt | date:'medium' }}
            <span *ngIf="build.duration"> &middot; {{ build.duration }}</span>
          </div>

          <div *ngIf="expandedId === build.id" class="build-card__logs">
            <pre *ngIf="logs.length > 0">{{ logs.join('\\n') }}</pre>
            <p *ngIf="logs.length === 0" class="text-muted">No logs available</p>
          </div>
        </div>

        <div *ngIf="builds.length === 0" class="builds__empty text-muted">
          No builds found. Builds will appear here from Forgejo CI.
        </div>
      </div>
    </div>
  `,
  styles: [`
    .builds {
      padding: var(--spacing-md);
    }

    .builds__header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: var(--spacing-md);
    }

    .builds__list {
      display: flex;
      flex-direction: column;
      gap: var(--spacing-sm);
    }

    .build-card {
      background: var(--bg-secondary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      padding: var(--spacing-md);
      cursor: pointer;
      transition: border-color 0.15s;

      &:hover {
        border-color: var(--text-muted);
      }
    }

    .build-card__header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: var(--spacing-xs);
    }

    .build-card__info {
      display: flex;
      gap: var(--spacing-sm);
      align-items: center;
    }

    .build-card__repo {
      font-weight: 600;
    }

    .build-card__branch {
      font-size: 12px;
    }

    .build-card__meta {
      font-size: 12px;
    }

    .build-card__logs {
      margin-top: var(--spacing-md);
      border-top: 1px solid var(--border-color);
      padding-top: var(--spacing-md);
    }

    .build-card__logs pre {
      font-size: 12px;
      max-height: 300px;
      overflow-y: auto;
    }

    .builds__empty {
      text-align: center;
      padding: var(--spacing-xl);
    }
  `]
})
export class BuildComponent implements OnInit, OnDestroy {
  builds: Build[] = [];
  expandedId = '';
  logs: string[] = [];

  private sub: Subscription | null = null;

  constructor(
    private wails: WailsService,
    private wsService: WebSocketService
  ) {}

  ngOnInit(): void {
    this.refresh();
    this.wsService.connect();
    this.sub = this.wsService.subscribe('build:status').subscribe(
      (msg: WSMessage) => {
        if (msg.data && typeof msg.data === 'object') {
          const update = msg.data as Build;
          const idx = this.builds.findIndex(b => b.id === update.id);
          if (idx >= 0) {
            this.builds[idx] = { ...this.builds[idx], ...update };
          } else {
            this.builds.unshift(update);
          }
        }
      }
    );
  }

  ngOnDestroy(): void {
    this.sub?.unsubscribe();
  }

  async refresh(): Promise<void> {
    this.builds = await this.wails.getBuilds();
  }

  async toggle(buildId: string): Promise<void> {
    if (this.expandedId === buildId) {
      this.expandedId = '';
      this.logs = [];
      return;
    }
    this.expandedId = buildId;
    this.logs = await this.wails.getBuildLogs(buildId);
  }

  trackBuild(_: number, build: Build): string {
    return build.id;
  }

  statusBadge(status: string): string {
    switch (status) {
      case 'success': return 'badge--success';
      case 'running': return 'badge--info';
      case 'failed': return 'badge--danger';
      default: return 'badge--warning';
    }
  }
}
