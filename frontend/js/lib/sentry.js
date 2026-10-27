/**
 * Browser Sentry bootstrap (TP9).
 * Front DSN is not a secret — it only allows sending events, not reading them.
 * The DSN only allows sending events.
 */
import * as Sentry from 'https://cdn.jsdelivr.net/npm/@sentry/browser@8.55.0/+esm'

/** Browser project DSN (ucc-mk, project 4512325842436096). Ingestion only. */
export const SENTRY_DSN = 'https://5f558a90e8705bec91166f16062abd92@o4512325828018176.ingest.us.sentry.io/4512325842436096'

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
