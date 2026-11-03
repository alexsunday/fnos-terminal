import { createApp } from 'vue';
import App from './App.vue';

export function bootstrap(): void {
  const app = createApp(App);
  app.mount('#app');
}

// Explicit startup call
bootstrap();
