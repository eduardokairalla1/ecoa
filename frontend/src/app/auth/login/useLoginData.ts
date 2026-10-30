/**
 * The login screen's data: signing in, and where to go next.
 */

// --- IMPORTS ---
import { useCallback, useEffect } from 'react'
import { useSubmit } from '@/app/forms/useSubmit'
import { useGoTo, useRouteState } from '@/app/navigation/navigation'
import { HttpError } from '@/logic/errors/httpError'
import { login, resendVerification } from '@/logic/datasource/auth'
import { settleSignOut, signedIn } from '@/logic/session/session'

// --- CODE ---
/**
 * Signs in. Where it lands is GuestOnly's call; the one detour taken here
 * is for an email that was never confirmed.
 *
 * @returns The submit runner, and what the previous screen handed over.
 */
export function useLoginData() {
  const goTo = useGoTo()
  const state = useRouteState<'login'>()

  // the notice says why the session ended; the next visit should not
  useEffect(() => {
    if (state.ended) settleSignOut()
  }, [state.ended])

  const action = useCallback(
    async (email: string, password: string) => {
      try {
        signedIn(await login(email, password))
      } catch (error) {
        // the account exists but never confirmed: finish that first, with a
        // fresh code, since the one from sign-up may have expired
        if (
          error instanceof HttpError &&
          error.code === 'email_not_verified_error'
        ) {
          const sent = await resendVerification(email).then(
            () => true,
            () => false,
          )
          goTo('verify', { email, unverified: true, sent })
          return
        }
        throw error
      }
    },
    [goTo],
  )

  return { submit: useSubmit(action), state }
}
