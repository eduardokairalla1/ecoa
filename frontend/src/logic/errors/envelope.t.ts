/**
 * Types for the backend's error envelope.
 */

// --- GLOBALS ---
/** What the backend answers a failure with: a slug and a message. */
export type ErrorEnvelope = {
  error: string
  message: string | string[]
}
