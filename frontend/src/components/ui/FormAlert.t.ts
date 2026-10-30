/**
 * Types for the FormAlert.
 */

// --- IMPORTS ---
import type { ReactNode } from 'react'

// --- GLOBALS ---
export type FormAlertTone = 'danger' | 'valid'

export type FormAlertProps = {
  /** Nothing renders while it is empty. */
  children?: ReactNode
  tone?: FormAlertTone
}
