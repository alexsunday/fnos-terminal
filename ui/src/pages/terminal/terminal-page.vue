<template>
  <div class="terminal-page" :style="{ backgroundColor: activeThemeColors.background }">
    <terminal-toolbar
      :connection-state="connectionState"
      :font-size="config.fontSize"
      :current-theme="config.theme"
      @reconnect="connectWebSocket"
      @clear="clearTerminal"
      @update:font-size="updateFontSize"
      @update:current-theme="updateTheme"
    />

    <div class="terminal-body">
      <div ref="terminalContainer" class="terminal-container"></div>

      <!-- Session Ended / Disconnected Action Bar -->
      <transition name="slide-fade">
        <div v-if="isSessionEnded" class="session-ended-banner">
          <div class="banner-info">
            <span class="banner-dot"></span>
            <span class="banner-text">会话已结束（按 Enter 键重新连接）</span>
          </div>
          <div class="banner-actions">
            <button class="banner-btn primary" title="按回车键重新连接" @click="connectWebSocket">
              <svg viewBox="0 0 24 24" class="btn-icon">
                <path fill="currentColor" d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
              </svg>
              <span>重新连接</span>
            </button>
            <button class="banner-btn secondary" title="关闭当前窗口" @click="closeCurrentWindow">
              <span>关闭窗口</span>
            </button>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import '@xterm/xterm/css/xterm.css';

import TerminalToolbar from './components/terminal-toolbar.vue';
import {
  ConnectionState,
  TerminalConfig,
  TerminalTheme,
  THEMES,
  ThemeColors,
  ResizePayload
} from './types';

export default defineComponent({
  name: 'TerminalPage',
  components: {
    TerminalToolbar
  },
  data() {
    return {
      connectionState: 'connecting' as ConnectionState,
      isSessionEnded: false,
      config: {
        fontSize: 14,
        theme: 'default' as TerminalTheme,
        cursorBlink: true,
        fontFamily: 'Menlo, Monaco, "Courier New", monospace'
      } as TerminalConfig,
      term: null as Terminal | null,
      fitAddon: null as FitAddon | null,
      ws: null as WebSocket | null,
      resizeObserver: null as ResizeObserver | null,
      pingTimer: null as number | null
    };
  },
  computed: {
    activeThemeColors(): ThemeColors {
      return THEMES[this.config.theme] || THEMES.default;
    }
  },
  mounted() {
    this.initTerminal();
    this.connectWebSocket();
    this.initResizeObserver();
  },
  beforeUnmount() {
    this.cleanup();
  },
  methods: {
    initTerminal(): void {
      const container = this.$refs.terminalContainer as HTMLElement;
      if (!container) return;

      const term = new Terminal({
        fontSize: this.config.fontSize,
        fontFamily: this.config.fontFamily,
        cursorBlink: this.config.cursorBlink,
        theme: this.activeThemeColors,
        allowProposedApi: true
      });

      const fitAddon = new FitAddon();
      const webLinksAddon = new WebLinksAddon();

      term.loadAddon(fitAddon);
      term.loadAddon(webLinksAddon);
      term.open(container);

      fitAddon.fit();

      term.onData((data: string) => {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
          this.ws.send(data);
        } else if (this.isSessionEnded || this.connectionState === 'disconnected') {
          // If session is ended and user presses Enter or Space, reconnect immediately!
          if (data === '\r' || data === '\n' || data === ' ') {
            this.connectWebSocket();
          }
        }
      });

      term.onResize((size: { cols: number; rows: number }) => {
        this.sendResize(size.cols, size.rows);
      });

      this.term = term;
      this.fitAddon = fitAddon;
    },

    connectWebSocket(): void {
      if (this.ws) {
        this.ws.close();
        this.ws = null;
      }

      this.isSessionEnded = false;
      this.connectionState = 'connecting';

      const loc = window.location;
      const wsProtocol = loc.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsHost = loc.host;
      const initialCols = this.term ? this.term.cols : 80;
      const initialRows = this.term ? this.term.rows : 24;

      const wsUrl = `${wsProtocol}//${wsHost}/app/terminal/ws?cols=${initialCols}&rows=${initialRows}`;

      try {
        const ws = new WebSocket(wsUrl);
        ws.binaryType = 'arraybuffer';

        ws.onopen = () => {
          this.connectionState = 'connected';
          this.isSessionEnded = false;
          if (this.fitAddon && this.term) {
            this.fitAddon.fit();
            this.sendResize(this.term.cols, this.term.rows);
          }
          this.startHeartbeat();
        };

        ws.onmessage = (event: MessageEvent) => {
          if (typeof event.data === 'string') {
            if (event.data.startsWith('{')) {
              try {
                const parsed = JSON.parse(event.data);
                if (parsed.type === 'pong') {
                  return;
                }
                if (parsed.type === 'exit') {
                  this.handleSessionExit();
                  return;
                }
              } catch {
                // not control JSON, write to terminal
              }
            }
            if (this.term) {
              this.term.write(event.data);
            }
          } else if (event.data instanceof ArrayBuffer) {
            if (this.term) {
              this.term.write(new Uint8Array(event.data));
            }
          }
        };

        ws.onerror = () => {
          this.connectionState = 'error';
        };

        ws.onclose = () => {
          this.connectionState = 'disconnected';
          this.stopHeartbeat();
          if (!this.isSessionEnded) {
            this.handleSessionExit();
          }
        };

        this.ws = ws;
      } catch (err: unknown) {
        this.connectionState = 'error';
      }
    },

    handleSessionExit(): void {
      if (this.isSessionEnded) return;
      this.isSessionEnded = true;
      this.connectionState = 'disconnected';
      if (this.term) {
        this.term.write('\r\n\x1b[90m────────────────────────────────────────────────────\x1b[0m\r\n\x1b[33;1m[会话已结束] 终端进程已退出 (按回车键重新连接)\x1b[0m\r\n');
      }
      this.stopHeartbeat();
    },

    closeCurrentWindow(): void {
      // 1. Try standard window.close()
      window.close();

      // 2. Dispatch close events to parent in case running inside fnOS desktop iframe
      try {
        if (window.parent && window.parent !== window) {
          window.parent.postMessage({ type: 'closeApp', appName: 'terminal' }, '*');
          window.parent.postMessage({ action: 'close' }, '*');
          window.parent.postMessage('close', '*');
        }
      } catch {
        // ignore cross-origin error
      }
    },

    sendResize(cols: number, rows: number): void {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        const payload: ResizePayload = {
          type: 'resize',
          cols,
          rows
        };
        this.ws.send(JSON.stringify(payload));
      }
    },

    initResizeObserver(): void {
      const container = this.$refs.terminalContainer as HTMLElement;
      if (!container || typeof ResizeObserver === 'undefined') return;

      this.resizeObserver = new ResizeObserver(() => {
        if (this.fitAddon && this.term) {
          try {
            this.fitAddon.fit();
          } catch {
            // container might be hidden during layout
          }
        }
      });

      this.resizeObserver.observe(container);
    },

    clearTerminal(): void {
      if (this.term) {
        this.term.clear();
      }
    },

    updateFontSize(newSize: number): void {
      this.config.fontSize = newSize;
      if (this.term) {
        this.term.options.fontSize = newSize;
        if (this.fitAddon) {
          this.fitAddon.fit();
        }
      }
    },

    updateTheme(newTheme: TerminalTheme): void {
      this.config.theme = newTheme;
      if (this.term) {
        this.term.options.theme = this.activeThemeColors;
      }
    },

    startHeartbeat(): void {
      this.stopHeartbeat();
      this.pingTimer = window.setInterval(() => {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
          this.ws.send(JSON.stringify({ type: 'ping' }));
        }
      }, 30000);
    },

    stopHeartbeat(): void {
      if (this.pingTimer !== null) {
        window.clearInterval(this.pingTimer);
        this.pingTimer = null;
      }
    },

    cleanup(): void {
      this.stopHeartbeat();
      if (this.resizeObserver) {
        this.resizeObserver.disconnect();
        this.resizeObserver = null;
      }
      if (this.ws) {
        this.ws.close();
        this.ws = null;
      }
      if (this.term) {
        this.term.dispose();
        this.term = null;
      }
    }
  }
});
</script>

<style scoped>
.terminal-page {
  display: flex;
  flex-direction: column;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  position: relative;
  transition: background-color 0.2s ease;
}

.terminal-body {
  flex: 1;
  position: relative;
  overflow: hidden;
  padding: 6px 8px 8px 8px;
}

.terminal-container {
  width: 100%;
  height: 100%;
  overflow: hidden;
}

:deep(.xterm) {
  height: 100%;
  padding: 2px;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;
}

/* Floating Session Ended Banner */
.session-ended-banner {
  position: absolute;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background-color: rgba(32, 32, 36, 0.94);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(82, 82, 91, 0.6);
  border-radius: 8px;
  padding: 8px 16px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
  z-index: 100;
  max-width: 90%;
  user-select: none;
}

.banner-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #e4e4e7;
  font-weight: 500;
}

.banner-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #eab308;
  box-shadow: 0 0 6px rgba(234, 179, 8, 0.6);
}

.banner-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.banner-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.banner-btn.primary {
  background-color: #2563eb;
  color: #ffffff;
  border-color: #3b82f6;
}

.banner-btn.primary:hover {
  background-color: #1d4ed8;
}

.banner-btn.secondary {
  background-color: #27272a;
  color: #d4d4d8;
  border-color: #3f3f46;
}

.banner-btn.secondary:hover {
  background-color: #3f3f46;
  color: #ffffff;
}

.btn-icon {
  width: 14px;
  height: 14px;
}

.slide-fade-enter-active {
  transition: all 0.25s ease-out;
}

.slide-fade-leave-active {
  transition: all 0.2s cubic-bezier(1, 0.5, 0.8, 1);
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  transform: translate(-50%, 15px);
  opacity: 0;
}
</style>
