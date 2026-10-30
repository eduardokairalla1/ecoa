/**
 * The welcome screen's data: choosing the handle.
 */

// --- IMPORTS ---
import { useCallback } from 'react'

import { useSubmit } from '@/app/forms/useSubmit'
import { HttpError } from '@/logic/errors/httpError'
import { createProfile, me } from '@/logic/datasource/auth'
import { signedIn, useSession } from '@/logic/session/session'

// --- CODE ---
/**
 * Chooses the handle. RequireAuth moves on once the account has one.
 *
 * @returns The submit runner, and the signed-in address.
 */
export function useChooseHandleData() {
  const session = useSession()

  const action = useCallback(async (handle: string) => {
    try {
      signedIn(await createProfile(handle))
    } catch (error) {
      // chosen elsewhere already: read the account, and the guard moves on
      if (error instanceof HttpError && error.code === 'profile_exists_error') {
        signedIn(await me())
        return
      }
      throw error
    }
  }, [])

  return {
    submit: useSubmit(action),
    email: session.status === 'signed-in' ? session.account.email : '',
  }
}
