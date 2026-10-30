/**
 * Types for the session store.
 */

// --- IMPORTS ---
import type { Account } from '@/logic/datasource/auth.t'

// --- GLOBALS ---
export type SignOutReason = 'logout' | 'deleted'

export type Session =
  | { status: 'loading' }
  | {
      status: 'signed-out'
      reason?: SignOutReason
    }
  | { status: 'signed-in'; account: Account }
  | { status: 'unavailable' }
