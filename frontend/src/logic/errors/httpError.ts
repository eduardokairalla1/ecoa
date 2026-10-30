/**
 * The error every request to the backend fails with.
 */

// --- CODE ---
/**
 * A failure the backend reported, or a request that never got an answer.
 */
export class HttpError extends Error {
  readonly status: number
  readonly code: string
  readonly fields: Record<string, string>

  /**
   * Builds the error from what the backend said.
   *
   * @param status - The HTTP status, 0 when nothing answered.
   * @param code - The backend's slug, e.g. `handle_taken_error`.
   * @param message - The backend's message, for logs only.
   * @param fields - Per-field reasons, for a request that did not validate.
   */
  constructor(
    status: number,
    code: string,
    message: string,
    fields: Record<string, string> = {},
  ) {
    super(message)
    this.name = 'HttpError'
    this.status = status
    this.code = code
    this.fields = fields
  }
}
