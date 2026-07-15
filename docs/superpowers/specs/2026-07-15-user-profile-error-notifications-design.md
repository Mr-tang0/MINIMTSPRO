# MINIMTS User Profile and Error Notifications

## Goal

Replace the static sidebar identity and error badge in `MINIMTS.vue` with the logged-in user and live system messages. Add a hover notification panel and a user profile modal without introducing global state or changing the existing login API contract.

## User Experience

- The sidebar avatar remains the existing `user.png` asset.
- `logo-text` displays the current user's username, loaded through `LoginService.Login('__last_login__', ...)`.
- The error badge displays the number of active system messages. It is hidden when the count is zero.
- Hovering the avatar area opens `Message.vue` immediately to its right. The panel lists each message with its arrival time and includes a clear button.
- Clicking the avatar area opens `User.vue` as a centered modal. It shows the avatar, username, email, role, registration/login time, and a logout button.
- Clicking outside the profile modal or its close button dismisses it. Logout closes the MINIMTS window and invokes the existing login-window method.

## Architecture and Data Flow

`MINIMTS.vue` owns:

- `currentUser`, populated on mount from the last-login service response.
- `systemMessages`, an array of `{ id, text, time }` records.
- `showUserModal`, controlling `User.vue`.

`MINIMTS.vue` listens to the existing Wails `system_message` event. Each non-empty payload is normalized to a display string and appended to `systemMessages`. The badge is derived from `systemMessages.length`. Clearing the list is a local state update; no backend persistence is required.

`Message.vue` receives the message list and emits `clear`. `User.vue` receives the user and emits `close` and `logout`. Both components remain presentational and scoped-styled.

## Error Handling

- If last-login lookup fails or returns no username, display `MTS` as the identity fallback.
- Empty system messages are ignored.
- If logout cannot close the Wails window, fall back to the browser window close behavior already used by `Login.vue`.
- Missing profile fields render a neutral `-` placeholder.

## Accessibility and Layout

- The avatar control is a button with an accessible label.
- The notification panel and modal use semantic headings and buttons.
- The modal overlay uses a strong scrim and traps no new external state; Escape closes the profile modal.
- The panel is positioned to the right of the 90px sidebar and stays within the viewport on narrow widths.
- Existing dark dashboard tokens are reused; new controls keep the current compact desktop-tool density.

## Verification

- Build the frontend with the repository's existing package scripts.
- Run Go tests that do not require hardware.
- Manually verify username loading, event-driven badge updates, clear behavior, profile modal display, Escape/outside-click close, and logout fallback.
