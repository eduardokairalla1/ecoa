/**
 * The events API: where a signed envelope goes to the network.
 */

// --- IMPORTS ---
import type { PublishResult } from '@/logic/datasource/events.t'
import { BASE } from '@/logic/datasource/http'
import { errorOf } from '@/logic/errors/envelope'
import { HttpError } from '@/logic/errors/httpError'
import type { Envelope } from '@/logic/protocol/envelope.t'

// --- GLOBALS ---
// the statuses that carry a PublishResult; a rejection is one of them
const RESULT_STATUSES = [200, 202, 400]

const OUTCOMES = ['accepted', 'already_exists', 'orphan', 'rejected']

// --- CODE ---
/**
 * Checks that a body is a PublishResult.
 *
 * @param body - The parsed JSON.
 * @returns Whether it has the shape.
 */
function isResult(body: unknown): body is PublishResult {
  if (typeof body !== 'object' || body === null) return false

  const { outcome, reason } = body as Record<string, unknown>

  return (
    typeof outcome === 'string' &&
    OUTCOMES.includes(outcome) &&
    (reason === null || typeof reason === 'string')
  )
}

/**
 * Submits a signed event through the gateway.
 *
 * @param envelope - The signed envelope, sent as it is.
 * @returns What the node did with it; a rejection is a result, not an error.
 */
export async function publishEvent(envelope: Envelope): Promise<PublishResult> {
  let response: Response

  try {
    response = await fetch(`${BASE}/events`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(envelope),
    })
  } catch {
    throw new HttpError(0, 'network_error', 'The backend did not answer.')
  }

  if (!RESULT_STATUSES.includes(response.status)) throw await errorOf(response)

  const body: unknown = await response.json().catch(() => null)

  if (!isResult(body)) {
    throw new HttpError(response.status, 'unknown_error', 'Unreadable result.')
  }

  return body
}
