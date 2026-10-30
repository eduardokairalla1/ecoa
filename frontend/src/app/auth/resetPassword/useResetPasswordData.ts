/**
 * The reset-password screen's data: the new password, and a new code.
 */

// --- IMPORTS ---
import { useCallback } from 'react'

import { useSubmit } from '@/app/forms/useSubmit'
import { useGoTo, useRouteState } from '@/app/navigation/navigation'
import { forgotPassword, resetPassword } from '@/logic/datasource/auth'

// --- CODE ---
/**
 * Sets the new password, then sends the person to sign in with it.
 *
 * @returns The change and resend runners, and the address carried over.
 */
export function useResetPasswordData() {
  const goTo = useGoTo()
  const state = useRouteState<'resetPassword'>()

  const change = useCallback(
    async (email: string, code: string, password: string) => {
      await resetPassword(email, code, password)
      goTo('login', { email, reset: true })
    },
    [goTo],
  )

  const resend = useCallback(async (email: string) => {
    await forgotPassword(email)
  }, [])

  return { change: useSubmit(change), resend: useSubmit(resend), state }
}
