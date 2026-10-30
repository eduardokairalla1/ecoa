/**
 * The one door to the backend.
 */

// --- IMPORTS ---
import type { RequestOptions } from '@/logic/datasource/http.t'
import { errorOf } from '@/logic/errors/envelope'
import { HttpError } from '@/logic/errors/httpError'

// --- GLOBALS ---
const BASE = import.meta.env.VITE_API_URL ?? '/api'

// --- CODE ---
/**
 * Calls the backend, with the session cookies.
 *
 * @param path - The route under the API prefix, e.g. `/auth/me`.
 * @param options - Method and JSON body.
 * @returns The JSON answer, or undefined when there is none.
 */
export async function request<T>(
  path: string,
  { method = 'GET', body }: RequestOptions = {},
): Promise<T> {
  let response: Response

  try {
    response = await fetch(`${BASE}${path}`, {
      method,
      credentials: 'include',
      headers: body === undefined ? {} : { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new HttpError(0, 'network_error', 'The backend did not answer.')
  }

  if (response.ok) {
    const text = await response.text()
    return (text ? JSON.parse(text) : undefined) as T
  }

  throw await errorOf(response)
}
