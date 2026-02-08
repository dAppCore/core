import { Injectable, OnDestroy } from '@angular/core';
import { Subject, Observable } from 'rxjs';
import { filter } from 'rxjs/operators';

export interface WSMessage {
  type: string;
  channel?: string;
  processId?: string;
  data?: unknown;
  timestamp: string;
}

@Injectable({ providedIn: 'root' })
export class WebSocketService implements OnDestroy {
  private ws: WebSocket | null = null;
  private messages$ = new Subject<WSMessage>();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private url = 'ws://127.0.0.1:9877/ws';
  private connected = false;

  connect(url?: string): void {
    if (url) this.url = url;
    this.doConnect();
  }

  private doConnect(): void {
    if (this.ws) {
      this.ws.close();
    }

    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      this.connected = true;
      console.log('[WS] Connected');
    };

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const msg: WSMessage = JSON.parse(event.data);
        this.messages$.next(msg);
      } catch {
        console.warn('[WS] Failed to parse message');
      }
    };

    this.ws.onclose = () => {
      this.connected = false;
      console.log('[WS] Disconnected, reconnecting in 3s...');
      this.reconnectTimer = setTimeout(() => this.doConnect(), 3000);
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  subscribe(channel: string): Observable<WSMessage> {
    // Send subscribe command to hub
    this.send({ type: 'subscribe', data: channel, timestamp: new Date().toISOString() });
    return this.messages$.pipe(
      filter(msg => msg.channel === channel)
    );
  }

  unsubscribe(channel: string): void {
    this.send({ type: 'unsubscribe', data: channel, timestamp: new Date().toISOString() });
  }

  send(msg: WSMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  get isConnected(): boolean {
    return this.connected;
  }

  get allMessages$(): Observable<WSMessage> {
    return this.messages$.asObservable();
  }

  ngOnDestroy(): void {
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.ws?.close();
    this.messages$.complete();
  }
}
