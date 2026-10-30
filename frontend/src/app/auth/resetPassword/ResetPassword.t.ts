/**
 * Types for the ResetPassword screen.
 */

// --- GLOBALS ---
export type ResetPasswordErrors = {
  email?: string
  code?: string
  password?: string
  confirm?: string
}

export type ResetPasswordState = {
  email?: string
}
