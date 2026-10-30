/**
 * Reads the error envelope the backend answers a failure with.
 */

// --- IMPORTS ---
import type { ErrorEnvelope } from '@/logic/errors/envelope.t'
import { HttpError } from '@/logic/errors/httpError'

// --- CODE ---
/**
 * Reads the per-field reasons out of a validation failure.
 *
 * @param message - The envelope's message: "field: reason" entries.
 * @returns The reasons, by field name.
 */
function fieldsOf(message: ErrorEnvelope['message']) {
  if (!Array.isArray(message)) return {}

  return Object.fromEntries(
    message.map((entry) => {
      const [field, ...reason] = entry.split(': ')
      return [field, reason.join(': ')]
    }),
  )
}

/**
 * Turns a failed response into an HttpError.
 *
 * @param response - The response that was not ok.
 * @returns The error it describes.
 */
export async function errorOf(response: Response) {

  // try to parse the error from the response
  try {
    const { error, message } = (await response.json()) as ErrorEnvelope
    const text = Array.isArray(message) ? message.join('; ') : message

    return new HttpError(response.status, error, text, fieldsOf(message))

  // response is not valid JSON: return a generic error
  } catch {
    return new HttpError(response.status, 'unknown_error', response.statusText)
  }
}
