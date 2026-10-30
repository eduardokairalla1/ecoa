/**
 * The verify screen's data: confirming the email, and asking again.
 */

// --- IMPORTS ---
import { useCallback } from 'react'

import { useSubmit } from '@/app/forms/useSubmit'
import { useGoTo, useRouteState } from '@/app/navigation/navigation'
import { resendVerification, verify } from '@/logic/datasource/auth'

// --- CODE ---
/**
 * Confirms the email, then sends the person to sign in.
 *
 * @returns The confirm and resend runners, and what the previous screen
 * handed over.
 */
export function useVerifyData() {
  const goTo = useGoTo()
  const state = useRouteState<'verify'>()

  const confirm = useCallback(
    async (email: string, code: string) => {
      await verify(email, code)
      goTo('login', { email, verified: true })
    },
    [goTo],
  )

  const resend = useCallback(async (email: string) => {
    await resendVerification(email)
  }, [])

  return { confirm: useSubmit(confirm), resend: useSubmit(resend), state }
}
