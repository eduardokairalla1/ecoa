/**
 * Types for the CodeField.
 */

// --- GLOBALS ---
export type CodeFieldProps = {
  label: string
  /** Only digits, at most `length` of them. */
  value: string
  onValueChange: (value: string) => void
  /** How many digits the code has. */
  length?: number
  error?: string
  hint?: string
  autoFocus?: boolean
}
