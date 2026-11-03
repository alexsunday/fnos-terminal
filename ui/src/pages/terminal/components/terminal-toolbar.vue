<template>
  <header class="terminal-toolbar">
    <div class="toolbar-section left">
      <div class="status-indicator" :class="connectionState">
        <span class="status-dot"></span>
        <span class="status-text">{{ statusText }}</span>
      </div>

      <button
        v-if="connectionState !== 'connected'"
        class="tool-btn action-btn reconnect-btn"
        title="重新连接终端"
        @click="$emit('reconnect')"
      >
        <svg viewBox="0 0 24 24" class="btn-icon">
          <path fill="currentColor" d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
        </svg>
        <span>重连</span>
      </button>

      <button
        class="tool-btn"
        title="清空终端屏幕"
        @click="$emit('clear')"
      >
        <svg viewBox="0 0 24 24" class="btn-icon">
          <path fill="currentColor" d="M19 4h-3.5l-1-1h-5l-1 1H5v2h14M6 19a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7H6v12z"/>
        </svg>
        <span>清屏</span>
      </button>
    </div>

    <div class="toolbar-section right">
      <div class="control-group font-controls">
        <button
          class="icon-btn"
          title="缩小字体"
          :disabled="fontSize <= 12"
          @click="decreaseFontSize"
        >
          A-
        </button>
        <span class="font-size-label">{{ fontSize }}px</span>
        <button
          class="icon-btn"
          title="放大字体"
          :disabled="fontSize >= 24"
          @click="increaseFontSize"
        >
          A+
        </button>
      </div>

      <div class="control-group">
        <select
          class="theme-select"
          :value="currentTheme"
          title="选择终端主题"
          @change="onThemeChange"
        >
          <option value="default">Default Dark</option>
          <option value="dracula">Dracula</option>
          <option value="monokai">Monokai</option>
          <option value="solarized-dark">Solarized Dark</option>
        </select>
      </div>

      <button
        class="icon-btn"
        title="在新窗口打开"
        @click="openInNewTab"
      >
        <svg viewBox="0 0 24 24" class="btn-icon">
          <path fill="currentColor" d="M14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3m-2 16H5V5h7V3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7h-2v7z"/>
        </svg>
      </button>

      <button
        class="icon-btn"
        :title="isFullscreen ? '退出全屏' : '全屏展示'"
        @click="toggleFullscreen"
      >
        <svg v-if="!isFullscreen" viewBox="0 0 24 24" class="btn-icon">
          <path fill="currentColor" d="M5 5h5v2H7v3H5V5m9 0h5v5h-2V7h-3V5m3 9h2v5h-5v-2h3v-3M7 14v3h3v2H5v-5h2z"/>
        </svg>
        <svg v-else viewBox="0 0 24 24" class="btn-icon">
          <path fill="currentColor" d="M14 14h5v2h-3v3h-2v-5m-9 0h5v5H8v-3H5v-2m3-9v3H5V6h5v2H8m6 0V6h5v2h-3v2h-2V8z"/>
        </svg>
      </button>
    </div>
  </header>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import { ConnectionState, TerminalTheme } from '../types';

export default defineComponent({
  name: 'TerminalToolbar',
  props: {
    connectionState: {
      type: String as PropType<ConnectionState>,
      required: true
    },
    fontSize: {
      type: Number,
      required: true
    },
    currentTheme: {
      type: String as PropType<TerminalTheme>,
      required: true
    }
  },
  emits: ['reconnect', 'clear', 'update:fontSize', 'update:currentTheme'],
  data() {
    return {
      isFullscreen: false
    };
  },
  computed: {
    statusText(): string {
      switch (this.connectionState) {
        case 'connected':
          return '已连接';
        case 'connecting':
          return '连接中...';
        case 'disconnected':
          return '已断开';
        case 'error':
          return '连接错误';
        default:
          return '未连接';
      }
    }
  },
  methods: {
    decreaseFontSize(): void {
      if (this.fontSize > 12) {
        this.$emit('update:fontSize', this.fontSize - 1);
      }
    },
    increaseFontSize(): void {
      if (this.fontSize < 24) {
        this.$emit('update:fontSize', this.fontSize + 1);
      }
    },
    onThemeChange(event: Event): void {
      const target = event.target as HTMLSelectElement;
      if (target) {
        this.$emit('update:currentTheme', target.value as TerminalTheme);
      }
    },
    openInNewTab(): void {
      window.open(window.location.href, '_blank');
    },
    toggleFullscreen(): void {
      if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen().then(() => {
          this.isFullscreen = true;
        }).catch(() => {
          this.isFullscreen = false;
        });
      } else {
        document.exitFullscreen().then(() => {
          this.isFullscreen = false;
        }).catch(() => {
          this.isFullscreen = false;
        });
      }
    }
  }
});
</script>

<style scoped>
.terminal-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 38px;
  background-color: #202024;
  border-bottom: 1px solid #2e2e34;
  padding: 0 12px;
  font-size: 13px;
  user-select: none;
}

.toolbar-section {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  padding-right: 8px;
  border-right: 1px solid #323238;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #71717a;
}

.status-indicator.connected .status-dot {
  background-color: #22c55e;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.6);
}

.status-indicator.connecting .status-dot {
  background-color: #eab308;
  animation: pulse 1s infinite alternate;
}

.status-indicator.disconnected .status-dot {
  background-color: #71717a;
}

.status-indicator.error .status-dot {
  background-color: #ef4444;
  box-shadow: 0 0 6px rgba(239, 68, 68, 0.6);
}

@keyframes pulse {
  0% { opacity: 0.4; }
  100% { opacity: 1; }
}

.tool-btn,
.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  background-color: transparent;
  color: #d4d4d8;
  border: 1px solid transparent;
  border-radius: 4px;
  padding: 3px 7px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.tool-btn:hover,
.icon-btn:hover {
  background-color: #2d2d34;
  color: #ffffff;
}

.tool-btn:disabled,
.icon-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.reconnect-btn {
  background-color: #27272a;
  border-color: #3f3f46;
}

.reconnect-btn:hover {
  background-color: #3f3f46;
}

.btn-icon {
  width: 15px;
  height: 15px;
}

.control-group {
  display: flex;
  align-items: center;
  gap: 4px;
  background-color: #27272a;
  border: 1px solid #323238;
  border-radius: 4px;
  padding: 1px 3px;
}

.font-size-label {
  font-size: 11px;
  color: #a1a1aa;
  min-width: 28px;
  text-align: center;
}

.theme-select {
  background-color: transparent;
  color: #d4d4d8;
  border: none;
  font-size: 12px;
  padding: 3px 6px;
  outline: none;
  cursor: pointer;
}

.theme-select option {
  background-color: #202024;
  color: #f4f4f5;
}
</style>
