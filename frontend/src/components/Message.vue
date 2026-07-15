<template>
  <section class="message-panel" aria-label="消息通知">
    <header class="message-header">
      <div>
        <p class="eyebrow">System</p>
        <h3>消息通知</h3>
      </div>
      <span class="message-count" aria-label="消息数量">{{ safeMessages.length }}</span>
    </header>

    <p v-if="safeMessages.length === 0" class="empty-state">暂无系统消息</p>

    <ul v-else class="message-list" aria-live="polite">
      <li v-for="message in safeMessages" :key="message.renderKey" class="message-item">
        <p class="message-text">{{ message.text }}</p>
        <time class="message-time" :datetime="toDateTime(message.time)">{{ message.time }}</time>
      </li>
    </ul>

    <footer class="message-footer">
      <button
      class="clear-button"
      type="button"
      :disabled="safeMessages.length === 0"
      @click="emit('clear')"
    >
      <i class="ri-delete-bin-line" aria-hidden="true"></i>
      清除消息
      </button>
    </footer>
  </section>
</template>

<script setup>
import { computed } from 'vue'

const isMessageItem = (message) => (
  message !== null &&
  typeof message === 'object' &&
  Object.prototype.hasOwnProperty.call(message, 'id') &&
  typeof message.text === 'string' &&
  Object.prototype.hasOwnProperty.call(message, 'time') &&
  (typeof message.time === 'string' ||
    typeof message.time === 'number' ||
    message.time instanceof Date)
)

const props = defineProps({
  messages: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['clear'])

const safeMessages = computed(() => {
  if (!Array.isArray(props.messages)) return []

  return props.messages
    .filter(isMessageItem)
    .map((message, index) => ({
      ...message,
      renderKey: `${String(message.id ?? 'message')}-${index}`
    }))
})

const toDateTime = (value) => {
  const date = value instanceof Date ? value : new Date(value)
  return Number.isNaN(date.getTime()) ? undefined : date.toISOString()
}
</script>

<style scoped>
.message-panel {
  width: min(360px, calc(100vw - 32px));
  min-width: min(280px, calc(100vw - 32px));
  height: min(420px, calc(100vh - 32px));
  max-height: min(420px, calc(100vh - 32px));
  box-sizing: border-box;
  padding: 18px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--text-regular, #cbd5e1);
  background: var(--bg-panel, #273549);
  border: 1px solid var(--border-color, #475569);
  border-radius: 10px;
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.32);
  font-family: 'Inter', 'Segoe UI', system-ui, sans-serif;
}

.message-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--border-color, #475569);
}

.eyebrow {
  margin: 0 0 4px;
  color: var(--accent-blue, #3b82f6);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

h3 {
  margin: 0;
  color: var(--text-primary, #f8fafc);
  font-size: 16px;
  font-weight: 650;
}

.message-count {
  min-width: 24px;
  height: 24px;
  padding: 0 6px;
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: var(--accent-blue, #3b82f6);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
}

.empty-state {
  margin: 22px 0;
  color: var(--text-muted, #94a3b8);
  font-size: 13px;
  text-align: center;
}

.message-list {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 14px 0 18px;
  padding: 0;
  overflow-y: auto;
  list-style: none;
}

.message-item {
  padding: 11px 12px;
  background: var(--bg-card, #334155);
  border: 1px solid var(--border-color, #475569);
  border-left: 3px solid var(--accent-blue, #3b82f6);
  border-radius: 6px;
}

.message-text {
  margin: 0 0 7px;
  color: var(--text-primary, #f8fafc);
  font-size: 13px;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.message-time {
  display: block;
  color: var(--text-muted, #94a3b8);
  font-size: 11px;
}

.clear-button {
  width: 100%;
  min-height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 9px 12px;
  color: var(--text-regular, #cbd5e1);
  background: transparent;
  border: 1px solid var(--border-color, #475569);
  border-radius: 6px;
  font: inherit;
  font-size: 13px;
  cursor: pointer;
  transition: 150ms ease;
}

.message-footer {
  flex-shrink: 0;
}

.clear-button:hover:not(:disabled),
.clear-button:focus-visible:not(:disabled) {
  color: var(--text-primary, #f8fafc);
  background: rgba(59, 130, 246, 0.12);
  border-color: var(--accent-blue, #3b82f6);
  outline: none;
}

.clear-button:disabled {
  color: var(--text-muted, #94a3b8);
  cursor: not-allowed;
  opacity: 0.55;
}
</style>
