/**
 * Types for the Login screen.
 */

// --- GLOBALS ---
/** What another screen can hand the login on its way here. */
export type LoginState = {
  email?: string
  verified?: boolean
  reset?: boolean
  from?: string
  ended?: 'logout' | 'deleted'
}
