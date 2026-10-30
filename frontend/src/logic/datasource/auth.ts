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
 * Signs in. The session travels in cookies the page never sees.
 *
 * @param email - The account's address.
 * @param password - The account's password.
 * @returns The account signed in.
 */
export function login(email: string, password: string) {
  return request<Account>('/auth/login', {
    method: 'POST',
    body: { email, password },
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
 * Emails a code that lets the account choose a new password.
 *
 * @param email - The account's address.
 * @returns Nothing.
 */
export function forgotPassword(email: string) {
  return request<void>('/auth/forgot-password', {
    method: 'POST',
    body: { email },
  })
}

/**
 * Sets a new password with the code the account was sent.
 *
 * @param email - The account's address.
 * @param code - The 6-digit code.
 * @param password - The new password.
 * @returns Nothing.
 */
export function resetPassword(email: string, code: string, password: string) {
  return request<void>('/auth/reset-password', {
    method: 'POST',
    body: { email, code, password },
  })
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

/**
 * Deletes the signed-in account for good.
 *
 * @param password - The current password, asked again.
 * @returns Nothing.
 */
export function deleteAccount(password: string) {
  return request<void>('/auth/delete-account', {
    method: 'POST',
    body: { password },
  })
}
