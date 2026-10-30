/**
 * The auth API: accounts, sessions and passwords.
 */

// --- IMPORTS ---
import { request } from '@/logic/datasource/http'
import type { Account } from '@/logic/datasource/auth.t'

// --- CODE ---
/**
 * Reads the signed-in account.
 *
 * @returns The account.
 */
export function me() {
  return request<Account>('/auth/me')
}
