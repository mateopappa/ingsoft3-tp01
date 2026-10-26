/**
 * Browser Sentry bootstrap (TP9).
 * Front DSN is not a secret — it only allows sending events, not reading them.
 * Leave SENTRY_DSN empty (default) to keep the SDK idle locally / in CI.
 */
import * as Sentry from 'https://cdn.jsdelivr.net/npm/@sentry/browser@8.55.0/+esm'

/** @type {string} Replace with your Sentry front-project DSN after creating the project. */
export const SENTRY_DSN = ''

export function initSentry() {
  if (!SENTRY_DSN) {
    return
  }
  Sentry.init({
    dsn: SENTRY_DSN,
    environment: window.location.hostname,
    integrations: [Sentry.browserTracingIntegration()],
    tracesSampleRate: 1.0,
  })
}
