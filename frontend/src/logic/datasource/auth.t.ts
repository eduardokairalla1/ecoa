/**
 * Types for the auth API.
 */

// --- GLOBALS ---
/** An account as the backend returns it. */
export type Account = {
  sub: string
  email: string
  handle: string | null
}
