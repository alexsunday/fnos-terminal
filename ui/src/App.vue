<template>
  <terminal-page />
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import TerminalPage from './pages/terminal/terminal-page.vue';

export default defineComponent({
  name: 'App',
  components: {
    TerminalPage
  }
});
</script>
