/**
 * Types for the useSubmit hook.
 */

// --- GLOBALS ---
export type Submit<A extends unknown[]> = {
  run: (...args: A) => Promise<boolean>
  pending: boolean
  code: string
  message: string
  reset: () => void
}
