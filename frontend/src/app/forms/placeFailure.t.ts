/**
 * Types for placeFailure.
 */

// --- IMPORTS ---
import type { Submit } from '@/app/forms/useSubmit.t'

// --- GLOBALS ---
/** The part of a submit runner that says how it last failed. */
export type Failure = Pick<Submit<[]>, 'code' | 'message'>

/** The field each failure code is about, for the codes that are. */
export type FieldOf<E> = Partial<Record<string, keyof E>>
