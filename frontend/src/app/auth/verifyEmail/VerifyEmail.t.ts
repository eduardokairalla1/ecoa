/**
 * Types for the VerifyEmail screen.
 */

// --- GLOBALS ---
export type VerifyState = {
  email?: string
  unverified?: boolean
  sent?: boolean
}

export type VerifyErrors = {
  email?: string
  code?: string
}
