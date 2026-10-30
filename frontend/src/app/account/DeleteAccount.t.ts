/**
 * Types for the DeleteAccount dialog.
 */

// --- IMPORTS ---
import type { Submit } from '@/app/forms/useSubmit.t'

// --- GLOBALS ---
export type DeleteAccountProps = {
  handle: string
  remove: Submit<[password: string]>
}
