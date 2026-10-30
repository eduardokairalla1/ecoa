/**
 * The dashboard's data: the signed-in account.
 */

// --- IMPORTS ---
import { useSession } from '@/logic/session/session'

// --- CODE ---
/**
 * Reads the account the dashboard greets.
 *
 * @returns The account, or null until it is known.
 */
export function useDashboardData() {
  const session = useSession()

  return {
    account: session.status === 'signed-in' ? session.account : null,
  }
}
