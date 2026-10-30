/**
 * Types for the RequireAuth guard.
 */

// --- GLOBALS ---
export type RequireAuthProps = {
  /**
   * True for the screens that need a handle; false for the one screen that
   * chooses it, which a complete account has no reason to see.
   */
  profile: boolean
}
