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
 * Reads the signed-in account.
 *
 * @returns The account.
 */
export function me() {
  return request<Account>('/auth/me')
}
