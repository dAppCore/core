import { Injectable } from '@angular/core';

// Type-safe wrapper for Wails v3 Go service bindings.
// At runtime, `window.go.main.{ServiceName}.{Method}()` returns a Promise.

interface WailsGo {
  main: {
    IDEService: {
      GetConnectionStatus(): Promise<ConnectionStatus>;
      GetDashboard(): Promise<DashboardData>;
      ShowWindow(name: string): Promise<void>;
    };
    ChatService: {
      SendMessage(sessionId: string, message: string): Promise<boolean>;
      GetHistory(sessionId: string): Promise<ChatMessage[]>;
      ListSessions(): Promise<Session[]>;
      CreateSession(name: string): Promise<Session>;
      GetPlanStatus(sessionId: string): Promise<PlanStatus>;
    };
    BuildService: {
      GetBuilds(repo: string): Promise<Build[]>;
      GetBuildLogs(buildId: string): Promise<string[]>;
    };
  };
}

export interface ConnectionStatus {
  bridgeConnected: boolean;
  laravelUrl: string;
  wsClients: number;
  wsChannels: number;
}

export interface DashboardData {
  connection: ConnectionStatus;
}

export interface ChatMessage {
  role: string;
  content: string;
  timestamp: string;
}

export interface Session {
  id: string;
  name: string;
  status: string;
  createdAt: string;
}

export interface PlanStatus {
  sessionId: string;
  status: string;
  steps: PlanStep[];
}

export interface PlanStep {
  name: string;
  status: string;
}

export interface Build {
  id: string;
  repo: string;
  branch: string;
  status: string;
  duration?: string;
  startedAt: string;
}

declare global {
  interface Window {
    go: WailsGo;
  }
}

@Injectable({ providedIn: 'root' })
export class WailsService {
  private get ide() { return window.go?.main?.IDEService; }
  private get chat() { return window.go?.main?.ChatService; }
  private get build() { return window.go?.main?.BuildService; }

  // IDE
  getConnectionStatus(): Promise<ConnectionStatus> {
    return this.ide?.GetConnectionStatus() ?? Promise.resolve({
      bridgeConnected: false, laravelUrl: '', wsClients: 0, wsChannels: 0
    });
  }

  getDashboard(): Promise<DashboardData> {
    return this.ide?.GetDashboard() ?? Promise.resolve({
      connection: { bridgeConnected: false, laravelUrl: '', wsClients: 0, wsChannels: 0 }
    });
  }

  showWindow(name: string): Promise<void> {
    return this.ide?.ShowWindow(name) ?? Promise.resolve();
  }

  // Chat
  sendMessage(sessionId: string, message: string): Promise<boolean> {
    return this.chat?.SendMessage(sessionId, message) ?? Promise.resolve(false);
  }

  getHistory(sessionId: string): Promise<ChatMessage[]> {
    return this.chat?.GetHistory(sessionId) ?? Promise.resolve([]);
  }

  listSessions(): Promise<Session[]> {
    return this.chat?.ListSessions() ?? Promise.resolve([]);
  }

  createSession(name: string): Promise<Session> {
    return this.chat?.CreateSession(name) ?? Promise.resolve({
      id: '', name, status: 'offline', createdAt: ''
    });
  }

  getPlanStatus(sessionId: string): Promise<PlanStatus> {
    return this.chat?.GetPlanStatus(sessionId) ?? Promise.resolve({
      sessionId, status: 'offline', steps: []
    });
  }

  // Build
  getBuilds(repo: string = ''): Promise<Build[]> {
    return this.build?.GetBuilds(repo) ?? Promise.resolve([]);
  }

  getBuildLogs(buildId: string): Promise<string[]> {
    return this.build?.GetBuildLogs(buildId) ?? Promise.resolve([]);
  }
}
