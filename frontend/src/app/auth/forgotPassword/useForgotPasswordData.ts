/**
 * The forgot-password screen's data: asking for a reset code.
 */

// --- IMPORTS ---
import { useCallback } from 'react'

import { useSubmit } from '@/app/forms/useSubmit'
import { useGoTo, useRouteState } from '@/app/navigation/navigation'
import { forgotPassword } from '@/logic/datasource/auth'

// --- CODE ---
/**
 * Asks for a reset code, then moves to entering it.
 *
 * @returns The submit runner, and the address the sign-in had typed.
 */
export function useForgotPasswordData() {
  const goTo = useGoTo()
  const state = useRouteState<'forgotPassword'>()

  const action = useCallback(
    async (email: string) => {
      await forgotPassword(email)
      goTo('resetPassword', { email })
    },
    [goTo],
  )

  return { submit: useSubmit(action), state }
}
