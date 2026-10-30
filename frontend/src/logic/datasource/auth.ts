/**
 * The auth API: accounts, sessions and passwords.
 */

// --- IMPORTS ---
import { request } from '@/logic/datasource/http'
import type { Account } from '@/logic/datasource/auth.t'

// --- CODE ---
/**
 * Creates an account and emails its confirmation code.
 *
 * @param handle - The public name.
 * @param email - The address the account signs in with.
 * @param password - The account's password.
 * @returns The account created.
 */
export function register(handle: string, email: string, password: string) {
  return request<Account>('/auth/register', {
    method: 'POST',
    body: { handle, email, password },
  })
}

/**
 * Confirms an email with the code it was sent.
 *
 * @param email - The account's address.
 * @param code - The 6-digit code.
 * @returns Nothing.
 */
export function verify(email: string, code: string) {
  return request<void>('/auth/verify', {
    method: 'POST',
    body: { email, code },
  })
}

/**
 * Emails a new confirmation code; the previous one stops working.
 *
 * @param email - The account's address.
 * @returns Nothing.
 */
export function resendVerification(email: string) {
  return request<void>('/auth/verify/resend', {
    method: 'POST',
    body: { email },
  })
}

/**
 * Ends the session.
 *
 * @returns Nothing.
 */
export function logout() {
  return request<void>('/auth/logout', { method: 'POST', retry: false })
}

/**
 * Reads the signed-in account.
 *
 * @returns The account.
 */
export function me() {
  return request<Account>('/auth/me')
}

/**
 * Chooses the handle of an account that has none yet.
 *
 * @param handle - The public name, chosen once.
 * @returns The account, now with its handle.
 */
export function createProfile(handle: string) {
  return request<Account>('/auth/profile', {
    method: 'POST',
    body: { handle },
  })
}
