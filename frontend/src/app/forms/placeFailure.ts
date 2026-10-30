/**
 * Puts the backend's failure where the form shows it.
 */

// --- IMPORTS ---
import type { Failure, FieldOf } from '@/app/forms/placeFailure.t'

// --- CODE ---
/**
 * Merges the backend's last failure into the form's own errors: under the
 * field it is about when there is one, above the form otherwise.
 *
 * @param errors - The form's own errors, by field.
 * @param failure - The submit runner whose failure it is.
 * @param fieldOf - The field each failure code belongs to.
 * @returns The errors under each field, and the alert above the form.
 */
export function placeFailure<E extends object>(
  errors: E,
  failure: Failure,
  fieldOf: FieldOf<E>,
) {
  const field = fieldOf[failure.code]
  if (!field) return { errors, alert: failure.message }

  return { errors: { ...errors, [field]: failure.message }, alert: '' }
}
