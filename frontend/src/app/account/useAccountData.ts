/**
 * The account screen's data: who is signed in, signing out, deleting.
 */

// --- IMPORTS ---
import { useCallback } from 'react'

import { useSubmit } from '@/app/forms/useSubmit'
import { deleteAccount, logout } from '@/logic/datasource/auth'
import { signedOut, useSession } from '@/logic/session/session'

// --- CODE ---
/**
 * Reads the account and ends it. RequireAuth moves on once signed out.
 *
 * @returns The account, and the sign-out and delete runners.
 */
export function useAccountData() {
  const session = useSession()

  const leave = useCallback(async () => {
    await logout()
    signedOut('logout')
  }, [])

  const remove = useCallback(async (password: string) => {
    await deleteAccount(password)
    signedOut('deleted')
  }, [])

  return {
    account: session.status === 'signed-in' ? session.account : null,
    signOut: useSubmit(leave),
    remove: useSubmit(remove),
  }
}
