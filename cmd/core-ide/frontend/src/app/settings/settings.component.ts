import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="settings">
      <h2>Settings</h2>

      <div class="settings__section">
        <h3>Connection</h3>
        <div class="form-group">
          <label class="form-label">Laravel WebSocket URL</label>
          <input
            class="form-input"
            [(ngModel)]="laravelUrl"
            placeholder="ws://localhost:9876/ws"
          />
        </div>
        <div class="form-group">
          <label class="form-label">Workspace Root</label>
          <input
            class="form-input"
            [(ngModel)]="workspaceRoot"
            placeholder="/path/to/workspace"
          />
        </div>
      </div>

      <div class="settings__section">
        <h3>Appearance</h3>
        <div class="form-group">
          <label class="form-label">Theme</label>
          <select class="form-select" [(ngModel)]="theme">
            <option value="dark">Dark</option>
            <option value="light">Light</option>
          </select>
        </div>
      </div>

      <div class="settings__actions">
        <button class="btn btn--primary" (click)="save()">Save Settings</button>
      </div>
    </div>
  `,
  styles: [`
    .settings {
      padding: var(--spacing-lg);
      max-width: 500px;
    }

    .settings__section {
      margin-top: var(--spacing-lg);
      padding-top: var(--spacing-lg);
      border-top: 1px solid var(--border-color);

      &:first-of-type {
        margin-top: var(--spacing-md);
        padding-top: 0;
        border-top: none;
      }
    }

    .settings__actions {
      margin-top: var(--spacing-lg);
    }
  `]
})
export class SettingsComponent implements OnInit {
  laravelUrl = 'ws://localhost:9876/ws';
  workspaceRoot = '.';
  theme = 'dark';

  ngOnInit(): void {
    // Settings will be loaded from the Go backend
    const saved = localStorage.getItem('ide-settings');
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        this.laravelUrl = parsed.laravelUrl ?? this.laravelUrl;
        this.workspaceRoot = parsed.workspaceRoot ?? this.workspaceRoot;
        this.theme = parsed.theme ?? this.theme;
      } catch {
        // Ignore parse errors
      }
    }
  }

  save(): void {
    localStorage.setItem('ide-settings', JSON.stringify({
      laravelUrl: this.laravelUrl,
      workspaceRoot: this.workspaceRoot,
      theme: this.theme,
    }));

    if (this.theme === 'light') {
      document.documentElement.setAttribute('data-theme', 'light');
    } else {
      document.documentElement.removeAttribute('data-theme');
    }
  }
}
