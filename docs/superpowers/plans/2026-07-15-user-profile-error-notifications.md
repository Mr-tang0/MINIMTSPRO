# User Profile and Error Notifications Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show the logged-in user in the MINIMTS sidebar, collect live system messages into an error badge, and provide message/profile interactions.

**Architecture:** `MINIMTS.vue` owns the current user, message list, and modal visibility. It loads the user through the existing `LoginService` binding and listens to the existing `system_message` Wails event. `Message.vue` and `User.vue` are focused presentational children using props and emits; no new global store or backend API is needed.

**Tech Stack:** Vue 3 `<script setup>`, Wails v3 runtime events/window APIs, existing Remix Icon styles, Vite frontend build, Go tests.

---

### Task 1: Add the message panel component

**Files:**
- Create: `frontend/src/components/Message.vue`

- [ ] **Step 1: Write the component contract**

Define a `messages` prop containing `{ id, text, time }` entries and a `clear` emit. Render a heading, empty state, message rows, and a clear button that is disabled when there are no messages.

- [ ] **Step 2: Implement the focused panel**

Use a compact dark panel with `position: absolute` placement supplied by the parent, semantic `<section>`/`<h3>`, readable timestamps, and a clear action. Do not own message state in this component.

- [ ] **Step 3: Run the frontend build**

Run `npm.cmd run build` from `frontend`.

Expected: the new component compiles without Vue template or script errors.

### Task 2: Add the user profile modal component

**Files:**
- Create: `frontend/src/components/User.vue`

- [ ] **Step 1: Write the component contract**

Define a `user` prop and `close`/`logout` emits. Read fields using the Wails-generated JSON names supported by the existing code (`username`, `email`, `role`, `registered_at`, plus `login_time` fallback).

- [ ] **Step 2: Implement modal behavior**

Render the avatar in the upper center, username below it, profile fields with `-` fallback, close button, outside-overlay click handling, Escape handling, and a logout button. Keep the component presentational; the parent performs window operations.

- [ ] **Step 3: Run the frontend build**

Run `npm.cmd run build` from `frontend`.

Expected: the modal component compiles and its event handlers are valid.

### Task 3: Connect user and message state in MINIMTS

**Files:**
- Modify: `frontend/src/components/MINIMTS.vue`

- [ ] **Step 1: Add state and imports**

Import `onUnmounted`, `LoginService`, `Window`, `Events`, `Message`, and `User`. Add:

```js
const currentUser = reactive({ username: '', email: '', role: '', registered_at: '', login_time: '' })
const systemMessages = ref([])
const showMessagePanel = ref(false)
const showUserModal = ref(false)
let removeSystemMessageListener
```

- [ ] **Step 2: Replace static sidebar markup**

Make `.sidebar-header` a keyboard-accessible button-like control with click and key handling for the profile modal. Bind the username to `.logo-text`, render the badge only when `systemMessages.length > 0`, and place `<Message>` to the right of the avatar area while the pointer is over the header/panel region.

- [ ] **Step 3: Add event and user-loading helpers**

Normalize event payloads from either a string or `{ message, text, error }`, ignore empty values, and append `{ id: crypto.randomUUID?.() || Date.now(), text, time: new Date() }`. On mount call `LoginService.Login('__last_login__', '000000')`, copy the returned user fields, and register `Events.On('system_message', handler)`. Store the returned disposer if available.

- [ ] **Step 4: Add clear, modal, Escape, and logout handlers**

Clear by assigning `systemMessages.value = []`; close the profile modal on Escape or overlay click; on logout call `LoginService.CallMINIMTSWindow()` and then `Window.Close()` with `window.close()` fallback. Remove the event listener and Escape listener in `onUnmounted`.

- [ ] **Step 5: Add parent styles**

Add stable positioning/z-index for the message panel, keyboard focus styling for the avatar control, and hide the badge at zero. Reuse existing dark variables and keep the panel inside the viewport on widths below 420px.

- [ ] **Step 6: Run frontend verification**

Run `npm.cmd run build` from `frontend`, then run `git diff --check`.

Expected: build succeeds and diff check reports no whitespace errors.

### Task 4: Verify the integrated behavior

**Files:**
- Test: `frontend/src/components/MINIMTS.vue`
- Test: `frontend/src/components/Message.vue`
- Test: `frontend/src/components/User.vue`

- [ ] **Step 1: Run repository Go tests**

Run `go test ./...` from the repository root.

Expected: tests pass or any pre-existing hardware/CGO limitation is recorded without changing unrelated code.

- [ ] **Step 2: Perform the UI smoke check**

Launch the existing Wails dev command, log in, verify the username replaces `MTS`, emit or trigger a `system_message`, verify the badge and hover panel, clear messages, open/close the profile modal, and exercise logout.

- [ ] **Step 3: Review the final diff**

Run `git status --short` and `git diff -- frontend/src/components/MINIMTS.vue frontend/src/components/Message.vue frontend/src/components/User.vue`.

Expected: only the requested frontend files plus the implementation plan are changed by this task; unrelated pre-existing worktree changes remain untouched.
