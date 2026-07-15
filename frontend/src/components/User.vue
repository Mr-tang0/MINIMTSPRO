<template>
  <div
    class="user-modal"
    role="presentation"
    @click.self="closeModal"
  >
    <section
      ref="dialogRef"
      class="user-dialog"
      data-user-dialog
      role="dialog"
      aria-modal="true"
      aria-label="User profile"
      tabindex="-1"
    >
      <button
        ref="closeButtonRef"
        class="close-button"
        type="button"
        aria-label="Close user profile"
        title="Close"
        @click="closeModal"
      >
        <i class="ri-close-line" aria-hidden="true"></i>
      </button>

      <div class="user-identity">
        <img
          class="user-avatar"
          :src="userAvatar"
          :alt="`${displayName} avatar`"
        />
        <h2 class="user-name">{{ displayName }}</h2>
      </div>

      <dl class="user-details">
        <div class="detail-row">
          <dt>Email</dt>
          <dd>{{ displayValue(user?.email) }}</dd>
        </div>
        <div class="detail-row">
          <dt>Role</dt>
          <dd>{{ displayValue(user?.role) }}</dd>
        </div>
        <div class="detail-row">
          <dt>Registered at</dt>
          <dd>
            <time
              v-if="registeredAtDateTime"
              :datetime="registeredAtDateTime"
            >
              {{ user?.registered_at }}
            </time>
            <span v-else-if="hasValue(user?.registered_at)">{{ user.registered_at }}</span>
            <span v-else>-</span>
          </dd>
        </div>
        <div class="detail-row">
          <dt>Last login</dt>
          <dd>
            <time
              v-if="lastLoginDateTime"
              :datetime="lastLoginDateTime"
            >
              {{ lastLoginTime }}
            </time>
            <span v-else-if="hasValue(lastLoginTime)">{{ lastLoginTime }}</span>
            <span v-else>-</span>
          </dd>
        </div>
      </dl>

      <button class="logout-button" type="button" @click="emit('logout')">
        <i class="ri-logout-box-r-line" aria-hidden="true"></i>
        <span>Log out</span>
      </button>
    </section>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, onUnmounted, ref } from 'vue'
import userAvatar from '../res/user.png'

const props = defineProps({
  user: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['close', 'logout'])
const dialogRef = ref(null)
const closeButtonRef = ref(null)
let previouslyFocusedElement = null
let isClosing = false
let hasRestoredFocus = false

const hasValue = (value) => (
  value !== null && value !== undefined && String(value).trim() !== ''
)

const displayValue = (value) => (
  hasValue(value)
    ? String(value)
    : '-'
)

const toDateTime = (value) => {
  if (!hasValue(value)) return undefined

  const date = value instanceof Date ? value : new Date(value)
  return Number.isNaN(date.getTime()) ? undefined : date.toISOString()
}

const getFocusableElements = () => (
  dialogRef.value?.querySelectorAll(
    'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
  ) ?? []
)

const restoreFocus = () => {
  if (hasRestoredFocus) return

  const opener = previouslyFocusedElement
  const activeElement = document.activeElement

  if (
    opener?.isConnected &&
    (activeElement === document.body || dialogRef.value?.contains(activeElement))
  ) {
    hasRestoredFocus = true
    previouslyFocusedElement = null
    opener.focus({ preventScroll: true })
  }
}

const closeModal = async () => {
  if (isClosing) return

  isClosing = true
  emit('close')
  await nextTick()

  restoreFocus()
}

const isTopmostDialog = () => {
  const dialogs = document.querySelectorAll('[data-user-dialog]')
  const topmostDialog = typeof dialogs.at === 'function'
    ? dialogs.at(-1)
    : dialogs[dialogs.length - 1]

  return dialogRef.value === topmostDialog
}

const handleKeydown = (event) => {
  if (!isTopmostDialog()) return

  if (event.key === 'Escape') {
    closeModal()
    return
  }

  if (event.key !== 'Tab') return

  const focusableElements = [...getFocusableElements()]
  if (focusableElements.length === 0) return

  const firstElement = focusableElements[0]
  const lastElement = focusableElements[focusableElements.length - 1]
  const activeElement = document.activeElement

  if (!dialogRef.value?.contains(activeElement)) {
    event.preventDefault()
    ;(event.shiftKey ? lastElement : firstElement).focus()
    return
  }

  if (event.shiftKey && activeElement === firstElement) {
    event.preventDefault()
    lastElement.focus()
  } else if (!event.shiftKey && activeElement === lastElement) {
    event.preventDefault()
    firstElement.focus()
  }
}

const displayName = computed(() => displayValue(props.user?.username || props.user?.name))
const lastLoginTime = computed(() => props.user?.login_time || props.user?.last_login)
const registeredAtDateTime = computed(() => toDateTime(props.user?.registered_at))
const lastLoginDateTime = computed(() => toDateTime(lastLoginTime.value))

onMounted(async () => {
  previouslyFocusedElement = document.activeElement
  document.addEventListener('keydown', handleKeydown)
  await nextTick()
  closeButtonRef.value?.focus({ preventScroll: true })
})
onBeforeUnmount(() => {
  restoreFocus()
})
onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.user-modal {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  overflow-y: auto;
  background: rgba(2, 6, 23, 0.72);
  color: var(--text-regular, #cbd5e1);
  font-family: 'Inter', 'Segoe UI', system-ui, sans-serif;
}

.user-dialog {
  position: relative;
  width: min(420px, 100%);
  max-height: calc(100vh - 40px);
  box-sizing: border-box;
  overflow-y: auto;
  padding: 32px;
  background: var(--bg-panel, #273549);
  border: 1px solid var(--border-color, #475569);
  border-radius: 10px;
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.42);
}

.close-button {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 40px;
  height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  color: var(--text-muted, #94a3b8);
  background: transparent;
  border: 1px solid transparent;
  border-radius: 6px;
  font-size: 20px;
  cursor: pointer;
  transition: 150ms ease;
}

.close-button:hover,
.close-button:focus-visible {
  color: var(--text-primary, #f8fafc);
  background: rgba(255, 255, 255, 0.08);
  border-color: var(--border-color, #475569);
  outline: none;
}

.user-identity {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 24px 24px;
  text-align: center;
}

.user-avatar {
  width: 88px;
  height: 88px;
  object-fit: cover;
  background: #0f172a;
  border: 3px solid rgba(59, 130, 246, 0.7);
  border-radius: 50%;
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.28);
}

.user-name {
  max-width: 100%;
  margin: 16px 0 0;
  overflow-wrap: anywhere;
  color: var(--text-primary, #f8fafc);
  font-size: 20px;
  font-weight: 650;
  line-height: 1.3;
}

.user-details {
  margin: 0;
  border-top: 1px solid var(--border-color, #475569);
}

.detail-row {
  display: grid;
  grid-template-columns: minmax(92px, 0.4fr) minmax(0, 1fr);
  gap: 16px;
  padding: 14px 0;
  border-bottom: 1px solid rgba(71, 85, 105, 0.7);
}

dt,
dd {
  min-width: 0;
  margin: 0;
  font-size: 13px;
  line-height: 1.45;
}

dt {
  color: var(--text-muted, #94a3b8);
}

dd {
  overflow-wrap: anywhere;
  color: var(--text-primary, #f8fafc);
  text-align: right;
}

.logout-button {
  width: 100%;
  min-height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 24px;
  padding: 10px 16px;
  color: #fecaca;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.38);
  border-radius: 6px;
  font: inherit;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: 150ms ease;
}

.logout-button:hover,
.logout-button:focus-visible {
  color: #fff1f2;
  background: rgba(239, 68, 68, 0.22);
  border-color: #ef4444;
  outline: none;
}

@media (max-width: 480px) {
  .user-modal {
    align-items: flex-start;
    padding: 12px;
  }

  .user-dialog {
    max-height: calc(100vh - 24px);
    padding: 28px 20px 20px;
  }

  .user-identity {
    padding-bottom: 20px;
  }

  .detail-row {
    grid-template-columns: 1fr;
    gap: 4px;
    padding: 12px 0;
  }

  dd {
    text-align: left;
  }
}

@media (prefers-reduced-motion: reduce) {
  .close-button,
  .logout-button {
    transition: none;
  }
}
</style>
