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
        <span class="btn-text">重连</span>
      </button>

      <button
        class="tool-btn"
        title="清空终端屏幕"
        @click="$emit('clear')"
      >
        <svg viewBox="0 0 24 24" class="btn-icon">
          <path fill="currentColor" d="M19 4h-3.5l-1-1h-5l-1 1H5v2h14M6 19a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7H6v12z"/>
        </svg>
        <span class="btn-text">清屏</span>
      </button>
    </div>

    <div class="toolbar-section right">
      <!-- Mobile Bar Toggle Button -->
      <button
        class="tool-btn toggle-bar-btn"
        :class="{ active: isMobileBarVisible }"
        title="切换虚拟按键辅助栏 (ESC, TAB, CTRL, 方向键)"
        @click="$emit('toggle-mobile-bar')"
      >
        <svg viewBox="0 0 24 24" class="btn-icon">
          <path fill="currentColor" d="M20 5H4c-1.1 0-1.99.9-1.99 2L2 17c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm-9 3h2v2h-2V8zm0 3h2v2h-2v-2zM8 8h2v2H8V8zm0 3h2v2H8v-2zm-1 2H5v-2h2v2zm0-3H5V8h2v2zm9 7H8v-2h8v2zm0-4h-2v-2h2v2zm0-3h-2V8h2v2zm3 3h-2v-2h2v2zm0-3h-2V8h2v2z"/>
        </svg>
        <span class="btn-text mobile-hide">辅助键</span>
      </button>

      <!-- Font Controls (hidden on narrow screens) -->
      <div class="control-group font-controls mobile-hide">
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

      <!-- Theme Dropdown (Button-based, completely eliminates iOS Form Assistant ^ v buttons) -->
      <div class="control-group theme-group" ref="themeGroup">
        <button
          type="button"
          class="theme-dropdown-btn"
          title="选择终端主题"
          @click.stop="toggleThemeDropdown"
        >
          <span class="theme-name">{{ currentThemeLabel }}</span>
          <svg viewBox="0 0 24 24" class="dropdown-arrow">
            <path fill="currentColor" d="M7 10l5 5 5-5z"/>
          </svg>
        </button>

        <div v-if="isThemeDropdownOpen" class="theme-dropdown-menu">
          <button
            v-for="item in themeList"
            :key="item.value"
            type="button"
            class="theme-item"
            :class="{ active: currentTheme === item.value }"
            @click="selectTheme(item.value)"
          >
            {{ item.label }}
          </button>
        </div>
      </div>

      <!-- Popout to new tab -->
      <button
        class="icon-btn mobile-hide"
        title="在新窗口打开"
        @click="openInNewTab"
      >
        <svg viewBox="0 0 24 24" class="btn-icon">
          <path fill="currentColor" d="M14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3m-2 16H5V5h7V3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7h-2v7z"/>
        </svg>
      </button>

      <!-- Fullscreen Toggle -->
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
    },
    isMobileBarVisible: {
      type: Boolean,
      default: false
    }
  },
  emits: ['reconnect', 'clear', 'update:fontSize', 'update:currentTheme', 'toggle-mobile-bar'],
  data() {
    return {
      isFullscreen: false,
      isThemeDropdownOpen: false
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
    },
    themeList(): Array<{ value: TerminalTheme; label: string }> {
      return [
        { value: 'default', label: 'Default' },
        { value: 'dracula', label: 'Dracula' },
        { value: 'monokai', label: 'Monokai' },
        { value: 'solarized-dark', label: 'Solarized' }
      ];
    },
    currentThemeLabel(): string {
      const found = this.themeList.find(t => t.value === this.currentTheme);
      return found ? found.label : 'Default';
    }
  },
  mounted() {
    document.addEventListener('click', this.handleOutsideClick);
  },
  beforeUnmount() {
    document.removeEventListener('click', this.handleOutsideClick);
  },
  methods: {
    toggleThemeDropdown(): void {
      this.isThemeDropdownOpen = !this.isThemeDropdownOpen;
    },
    selectTheme(theme: TerminalTheme): void {
      this.$emit('update:currentTheme', theme);
      this.isThemeDropdownOpen = false;
    },
    handleOutsideClick(event: MouseEvent): void {
      const themeGroup = this.$refs.themeGroup as HTMLElement;
      if (themeGroup && !themeGroup.contains(event.target as Node)) {
        this.isThemeDropdownOpen = false;
      }
    },
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
  height: 40px;
  background-color: #202024;
  border-bottom: 1px solid #2e2e34;
  padding: 0 10px;
  font-size: 13px;
  user-select: none;
  flex-shrink: 0;
}

.toolbar-section {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  padding-right: 6px;
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
  padding: 4px 7px;
  font-size: 12px;
  cursor: pointer;
  touch-action: manipulation;
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

.toggle-bar-btn.active {
  background-color: #2563eb;
  color: #ffffff;
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

.theme-group {
  position: relative;
}

.theme-dropdown-btn {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  background-color: transparent;
  color: #d4d4d8;
  border: none;
  font-size: 12px;
  padding: 3px 6px;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.15s ease;
}

.theme-dropdown-btn:hover {
  background-color: #2d2d34;
  color: #ffffff;
}

.dropdown-arrow {
  width: 12px;
  height: 12px;
  color: #a1a1aa;
}

.theme-dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 4px;
  background-color: #202024;
  border: 1px solid #3f3f46;
  border-radius: 6px;
  padding: 4px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
  z-index: 100;
  min-width: 105px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.theme-item {
  display: block;
  width: 100%;
  text-align: left;
  background-color: transparent;
  color: #d4d4d8;
  border: none;
  border-radius: 4px;
  padding: 6px 10px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.theme-item:hover {
  background-color: #2d2d34;
  color: #ffffff;
}

.theme-item.active {
  background-color: #2563eb;
  color: #ffffff;
  font-weight: 500;
}

/* Responsive adjustments */
@media (max-width: 640px) {
  .mobile-hide {
    display: none !important;
  }
  .terminal-toolbar {
    padding: 0 6px;
  }
  .toolbar-section {
    gap: 4px;
  }
}

@media (max-width: 420px) {
  .status-text {
    display: none;
  }
  .status-indicator {
    padding-right: 4px;
    border-right: none;
  }
  .btn-text {
    display: none;
  }
}
</style>
