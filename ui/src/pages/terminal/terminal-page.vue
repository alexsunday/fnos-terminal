<template>
  <div
    class="terminal-page"
    :style="{
      backgroundColor: activeThemeColors.background,
      height: pageHeight
    }"
  >
    <terminal-toolbar
      :connection-state="connectionState"
      :font-size="config.fontSize"
      :current-theme="config.theme"
      :is-mobile-bar-visible="isMobileBarVisible"
      @reconnect="connectWebSocket"
      @clear="clearTerminal"
      @update:font-size="updateFontSize"
      @update:current-theme="updateTheme"
      @toggle-mobile-bar="toggleMobileBar"
    />

    <div class="terminal-body" @click="focusTerminal" @touchstart="focusTerminal">
      <div ref="terminalContainer" class="terminal-container"></div>

      <!-- Session Ended Floating Capsule (Dynamic Island Style) -->
      <transition name="capsule-pop">
        <div
          v-if="isSessionEnded"
          class="session-ended-capsule"
          :class="{ 'has-mobile-bar': isMobileBarVisible }"
          title="点击重新连接"
          @click="connectWebSocket"
        >
          <span class="capsule-dot"></span>
          <span class="capsule-title">会话已结束</span>
          <span class="capsule-desktop-hint">（按 Enter 重连）</span>
          <button class="capsule-btn" title="重新连接" @click.stop="connectWebSocket">
            <svg viewBox="0 0 24 24" class="capsule-btn-icon">
              <path fill="currentColor" d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
            </svg>
            <span>重新连接</span>
          </button>
        </div>
      </transition>
    </div>

    <!-- Mobile Virtual Accessory Toolbar -->
    <terminal-mobile-bar
      v-if="isMobileBarVisible"
      :is-ctrl-active="isCtrlActive"
      :is-alt-active="isAltActive"
      :is-keyboard-open="isKeyboardOpen"
      @send-key="handleMobileKey"
      @toggle-ctrl="toggleCtrl"
      @toggle-alt="toggleAlt"
      @focus-terminal="focusTerminal"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import '@xterm/xterm/css/xterm.css';

import TerminalToolbar from './components/terminal-toolbar.vue';
import TerminalMobileBar from './components/terminal-mobile-bar.vue';
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
    TerminalToolbar,
    TerminalMobileBar
  },
  data() {
    return {
      connectionState: 'connecting' as ConnectionState,
      isSessionEnded: false,
      isMobileBarVisible: false,
      isCtrlActive: false,
      isAltActive: false,
      pageHeight: '100%',
      isKeyboardOpen: false,
      viewportCleanup: null as (() => void) | null,
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
    this.detectMobileDevice();
    this.initTerminal();
    this.connectWebSocket();
    this.initResizeObserver();
    this.initViewportHandler();
  },
  beforeUnmount() {
    this.cleanup();
  },
  methods: {
    detectMobileDevice(): void {
      if (typeof window !== 'undefined') {
        const hasTouch = 'ontouchstart' in window || (navigator.maxTouchPoints && navigator.maxTouchPoints > 0);
        const isNarrow = window.innerWidth <= 768;
        this.isMobileBarVisible = Boolean(hasTouch || isNarrow);
      }
    },

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
          let payload = data;

          // Apply sticky Ctrl modifier if active
          if (this.isCtrlActive && payload.length === 1) {
            const charCode = payload.toUpperCase().charCodeAt(0);
            if (charCode >= 64 && charCode <= 95) {
              payload = String.fromCharCode(charCode - 64);
            }
            this.isCtrlActive = false;
          }

          // Apply sticky Alt modifier if active
          if (this.isAltActive && payload.length >= 1) {
            payload = '\x1b' + payload;
            this.isAltActive = false;
          }

          this.ws.send(payload);
        } else if (this.isSessionEnded || this.connectionState === 'disconnected') {
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

      const basePath = (loc.pathname ? loc.pathname.replace(/\/+$/, '') : '') || '/app/terminal';
      const wsUrl = `${wsProtocol}//${wsHost}${basePath}/ws?cols=${initialCols}&rows=${initialRows}`;

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

    handleMobileKey(key: string): void {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(key);
      }
      this.focusTerminal();
    },

    toggleCtrl(): void {
      this.isCtrlActive = !this.isCtrlActive;
      this.focusTerminal();
    },

    toggleAlt(): void {
      this.isAltActive = !this.isAltActive;
      this.focusTerminal();
    },

    toggleMobileBar(): void {
      this.isMobileBarVisible = !this.isMobileBarVisible;
      this.$nextTick(() => {
        if (this.fitAddon) {
          this.fitAddon.fit();
        }
      });
    },

    focusTerminal(): void {
      if (this.term) {
        this.term.focus();
        setTimeout(() => {
          if (window.scrollY !== 0) {
            window.scrollTo(0, 0);
          }
        }, 10);
      }
    },

    initViewportHandler(): void {
      if (typeof window === 'undefined') return;

      const handleViewport = () => {
        const vv = window.visualViewport;
        const windowH = window.innerHeight;

        if (vv) {
          // Virtual keyboard is active when visualViewport height is substantially smaller
          const isKeyboardActive = vv.height < windowH - 70;
          this.isKeyboardOpen = isKeyboardActive;

          if (isKeyboardActive) {
            this.pageHeight = `${vv.height}px`;
          } else {
            this.pageHeight = '100%';
            if (window.scrollY !== 0) {
              window.scrollTo(0, 0);
            }
          }
        }

        // Always prevent iOS window scroll displacement
        if (window.scrollY !== 0) {
          window.scrollTo(0, 0);
        }

        if (this.fitAddon && this.term) {
          try {
            this.fitAddon.fit();
            this.sendResize(this.term.cols, this.term.rows);
          } catch {
            // ignore layout error
          }
        }
      };

      if (window.visualViewport) {
        window.visualViewport.addEventListener('resize', handleViewport);
        window.visualViewport.addEventListener('scroll', handleViewport);
      } else {
        window.addEventListener('resize', handleViewport);
      }

      const handleScrollReset = () => {
        if (!this.isKeyboardOpen && window.scrollY !== 0) {
          window.scrollTo(0, 0);
        }
      };
      window.addEventListener('scroll', handleScrollReset);

      this.viewportCleanup = () => {
        if (window.visualViewport) {
          window.visualViewport.removeEventListener('resize', handleViewport);
          window.visualViewport.removeEventListener('scroll', handleViewport);
        } else {
          window.removeEventListener('resize', handleViewport);
        }
        window.removeEventListener('scroll', handleScrollReset);
      };
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
      if (this.viewportCleanup) {
        this.viewportCleanup();
        this.viewportCleanup = null;
      }
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
  height: 100%;
  overflow: hidden;
  position: relative;
  transition: background-color 0.2s ease;
}

.terminal-body {
  flex: 1;
  position: relative;
  overflow: hidden;
  padding: 4px 6px;
  cursor: text;
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

/* Floating Session Ended Capsule (Dynamic Island Style) */
.session-ended-capsule {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 5px 6px 5px 14px;
  background-color: rgba(24, 24, 27, 0.88);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 9999px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5), 0 2px 6px rgba(0, 0, 0, 0.25);
  z-index: 100;
  white-space: nowrap;
  user-select: none;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, bottom 0.2s ease;
}

.session-ended-capsule.has-mobile-bar {
  bottom: 12px;
}

.session-ended-capsule:hover {
  box-shadow: 0 12px 36px rgba(0, 0, 0, 0.6), 0 0 0 1px rgba(255, 255, 255, 0.25);
}

.session-ended-capsule:active {
  transform: translateX(-50%) scale(0.98);
}

.capsule-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #f59e0b;
  box-shadow: 0 0 8px rgba(245, 158, 11, 0.8);
  flex-shrink: 0;
  animation: pulse-dot 2s infinite ease-in-out;
}

@keyframes pulse-dot {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
    box-shadow: 0 0 8px rgba(245, 158, 11, 0.8);
  }
  50% {
    opacity: 0.55;
    transform: scale(0.85);
    box-shadow: 0 0 3px rgba(245, 158, 11, 0.4);
  }
}

.capsule-title {
  font-size: 13px;
  font-weight: 500;
  color: #f4f4f5;
  white-space: nowrap;
  flex-shrink: 0;
}

.capsule-desktop-hint {
  font-size: 12px;
  color: #a1a1aa;
  white-space: nowrap;
  flex-shrink: 0;
}

@media (max-width: 600px) {
  .capsule-desktop-hint {
    display: none;
  }
}

.capsule-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 12px;
  font-size: 12px;
  font-weight: 500;
  color: #ffffff;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 9999px;
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
  transition: all 0.15s ease;
  box-shadow: 0 2px 6px rgba(37, 99, 235, 0.35);
}

.capsule-btn:hover {
  background: linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%);
  box-shadow: 0 4px 10px rgba(37, 99, 235, 0.5);
}

.capsule-btn:active {
  transform: scale(0.96);
}

.capsule-btn-icon {
  width: 13px;
  height: 13px;
  flex-shrink: 0;
}

.capsule-pop-enter-active,
.capsule-pop-leave-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.capsule-pop-enter-from,
.capsule-pop-leave-to {
  opacity: 0;
  transform: translate(-50%, 16px) scale(0.92);
}
</style>
