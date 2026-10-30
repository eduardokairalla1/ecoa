/**
 * Types for the Login screen.
 */

// --- GLOBALS ---
export type LoginErrors = {
  email?: string
  password?: string
}

/** What another screen can hand the login on its way here. */
export type LoginState = {
  email?: string
  verified?: boolean
  reset?: boolean
  from?: string
  ended?: 'logout' | 'deleted'
}
