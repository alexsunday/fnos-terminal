<template>
  <div class="terminal-mobile-bar" :class="{ 'keyboard-open': isKeyboardOpen }">
    <div class="keys-scroll-container">
      <!-- ESC -->
      <button
        class="key-btn"
        title="Escape"
        @mousedown.prevent="sendKey('\x1b')"
        @touchstart.prevent="sendKey('\x1b')"
      >
        ESC
      </button>

      <!-- TAB -->
      <button
        class="key-btn"
        title="Tab (补全)"
        @mousedown.prevent="sendKey('\t')"
        @touchstart.prevent="sendKey('\t')"
      >
        TAB
      </button>

      <!-- CTRL Toggle -->
      <button
        class="key-btn toggle-btn"
        :class="{ active: isCtrlActive }"
        title="Control 键 (按下后下一个按键生效 Ctrl 组合)"
        @mousedown.prevent="$emit('toggle-ctrl')"
        @touchstart.prevent="$emit('toggle-ctrl')"
      >
        CTRL
      </button>

      <!-- ALT Toggle -->
      <button
        class="key-btn toggle-btn"
        :class="{ active: isAltActive }"
        title="Alt 键"
        @mousedown.prevent="$emit('toggle-alt')"
        @touchstart.prevent="$emit('toggle-alt')"
      >
        ALT
      </button>

      <!-- Arrow Left -->
      <button
        class="key-btn arrow-btn"
        title="左光标"
        @mousedown.prevent="sendKey('\x1b[D')"
        @touchstart.prevent="sendKey('\x1b[D')"
      >
        ←
      </button>

      <!-- Arrow Down -->
      <button
        class="key-btn arrow-btn"
        title="下光标 / 历史命令"
        @mousedown.prevent="sendKey('\x1b[B')"
        @touchstart.prevent="sendKey('\x1b[B')"
      >
        ↓
      </button>

      <!-- Arrow Up -->
      <button
        class="key-btn arrow-btn"
        title="上光标 / 历史命令"
        @mousedown.prevent="sendKey('\x1b[A')"
        @touchstart.prevent="sendKey('\x1b[A')"
      >
        ↑
      </button>

      <!-- Arrow Right -->
      <button
        class="key-btn arrow-btn"
        title="右光标"
        @mousedown.prevent="sendKey('\x1b[C')"
        @touchstart.prevent="sendKey('\x1b[C')"
      >
        →
      </button>

      <!-- Common Symbols -->
      <button
        class="key-btn symbol-btn"
        @mousedown.prevent="sendKey('/')"
        @touchstart.prevent="sendKey('/')"
      >
        /
      </button>
      <button
        class="key-btn symbol-btn"
        @mousedown.prevent="sendKey('-')"
        @touchstart.prevent="sendKey('-')"
      >
        -
      </button>
      <button
        class="key-btn symbol-btn"
        @mousedown.prevent="sendKey('|')"
        @touchstart.prevent="sendKey('|')"
      >
        |
      </button>
      <button
        class="key-btn symbol-btn"
        @mousedown.prevent="sendKey('~')"
        @touchstart.prevent="sendKey('~')"
      >
        ~
      </button>

      <!-- Keyboard Focus Button -->
      <button
        class="key-btn keyboard-btn"
        title="调出虚拟键盘"
        @mousedown.prevent="$emit('focus-terminal')"
        @touchstart.prevent="$emit('focus-terminal')"
      >
        <svg viewBox="0 0 24 24" class="btn-icon">
          <path fill="currentColor" d="M20 5H4c-1.1 0-1.99.9-1.99 2L2 17c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm-9 3h2v2h-2V8zm0 3h2v2h-2v-2zM8 8h2v2H8V8zm0 3h2v2H8v-2zm-1 2H5v-2h2v2zm0-3H5V8h2v2zm9 7H8v-2h8v2zm0-4h-2v-2h2v2zm0-3h-2V8h2v2zm3 3h-2v-2h2v2zm0-3h-2V8h2v2z"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'TerminalMobileBar',
  props: {
    isCtrlActive: {
      type: Boolean,
      default: false
    },
    isAltActive: {
      type: Boolean,
      default: false
    },
    isKeyboardOpen: {
      type: Boolean,
      default: false
    }
  },
  emits: ['send-key', 'toggle-ctrl', 'toggle-alt', 'focus-terminal'],
  methods: {
    sendKey(key: string): void {
      this.$emit('send-key', key);
    }
  }
});
</script>

<style scoped>
.terminal-mobile-bar {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  width: 100%;
  box-sizing: border-box;
  background-color: #1e1e22;
  border-top: 1px solid #2e2e34;
  padding-top: 4px;
  padding-left: 8px;
  padding-right: 8px;
  padding-bottom: max(6px, env(safe-area-inset-bottom, 6px));
  height: calc(44px + env(safe-area-inset-bottom, 0px));
  overflow-x: auto;
  overflow-y: hidden;
  user-select: none;
  -webkit-user-select: none;
  z-index: 50;
  transition: padding-bottom 0.15s ease, height 0.15s ease;
}

.terminal-mobile-bar.keyboard-open {
  height: 42px;
  padding-bottom: 4px;
}

.keys-scroll-container {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  overflow-x: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.keys-scroll-container::-webkit-scrollbar {
  display: none;
}

.key-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 34px;
  min-width: 40px;
  padding: 0 8px;
  background-color: #2a2a30;
  color: #e4e4e7;
  border: 1px solid #3f3f46;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  touch-action: manipulation;
  transition: all 0.1s ease;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
}

.key-btn:active {
  background-color: #3f3f46;
  transform: translateY(1px);
}

.toggle-btn.active {
  background-color: #2563eb;
  color: #ffffff;
  border-color: #60a5fa;
  box-shadow: 0 0 8px rgba(37, 99, 235, 0.6);
}

.arrow-btn {
  min-width: 36px;
  font-size: 15px;
}

.symbol-btn {
  min-width: 32px;
  font-family: monospace;
  font-size: 14px;
}

.keyboard-btn {
  min-width: 38px;
  background-color: #323238;
}

.btn-icon {
  width: 18px;
  height: 18px;
}
</style>
