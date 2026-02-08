import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { ChatComponent } from '../chat/chat.component';
import { BuildComponent } from '../build/build.component';
import { DashboardComponent } from '../dashboard/dashboard.component';
import { JellyfinComponent } from '../jellyfin/jellyfin.component';

type Panel = 'chat' | 'build' | 'dashboard' | 'jellyfin';

@Component({
  selector: 'app-main',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive, RouterOutlet, ChatComponent, BuildComponent, DashboardComponent, JellyfinComponent],
  template: `
    <div class="ide">
      <nav class="ide__sidebar">
        <div class="ide__logo">Core IDE</div>
        <ul class="ide__nav">
          <li
            *ngFor="let item of navItems"
            class="ide__nav-item"
            [class.active]="activePanel === item.id"
            (click)="activePanel = item.id"
          >
            <span class="ide__nav-icon">{{ item.icon }}</span>
            <span class="ide__nav-label">{{ item.label }}</span>
          </li>
        </ul>
        <div class="ide__nav-footer text-muted">v0.1.0</div>
      </nav>

      <main class="ide__content">
        <app-chat *ngIf="activePanel === 'chat'" />
        <app-build *ngIf="activePanel === 'build'" />
        <app-dashboard *ngIf="activePanel === 'dashboard'" />
        <app-jellyfin *ngIf="activePanel === 'jellyfin'" />
      </main>
    </div>
  `,
  styles: [`
    .ide {
      display: flex;
      height: 100vh;
      overflow: hidden;
    }

    .ide__sidebar {
      width: var(--sidebar-width);
      background: var(--bg-sidebar);
      border-right: 1px solid var(--border-color);
      display: flex;
      flex-direction: column;
      padding: var(--spacing-md) 0;
      flex-shrink: 0;
    }

    .ide__logo {
      padding: 0 var(--spacing-md);
      font-size: 16px;
      font-weight: 700;
      color: var(--accent-primary);
      margin-bottom: var(--spacing-lg);
    }

    .ide__nav {
      list-style: none;
      flex: 1;
    }

    .ide__nav-item {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);
      padding: var(--spacing-sm) var(--spacing-md);
      cursor: pointer;
      color: var(--text-secondary);
      transition: all 0.15s;
      border-left: 3px solid transparent;

      &:hover {
        color: var(--text-primary);
        background: var(--bg-tertiary);
      }

      &.active {
        color: var(--accent-primary);
        background: rgba(57, 208, 216, 0.08);
        border-left-color: var(--accent-primary);
      }
    }

    .ide__nav-icon {
      font-size: 16px;
      width: 20px;
      text-align: center;
    }

    .ide__nav-footer {
      padding: var(--spacing-sm) var(--spacing-md);
      font-size: 12px;
    }

    .ide__content {
      flex: 1;
      overflow: auto;
    }
  `]
})
export class MainComponent {
  activePanel: Panel = 'dashboard';

  navItems: { id: Panel; label: string; icon: string }[] = [
    { id: 'dashboard', label: 'Dashboard', icon: '\u25A6' },
    { id: 'chat', label: 'Chat', icon: '\u2709' },
    { id: 'build', label: 'Builds', icon: '\u2699' },
    { id: 'jellyfin', label: 'Jellyfin', icon: '\u25B6' },
  ];
}
