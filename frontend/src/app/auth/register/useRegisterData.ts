/**
 * The register screen's data: creating the account.
 */

// --- IMPORTS ---
import { useCallback } from 'react'

import { useSubmit } from '@/app/forms/useSubmit'
import { useGoTo } from '@/app/navigation/navigation'
import { register } from '@/logic/datasource/auth'

// --- CODE ---
/**
 * Creates the account, then moves to confirming its email.
 *
 * @returns The submit runner.
 */
export function useRegisterData() {
  const goTo = useGoTo()

  const action = useCallback(
    async (handle: string, email: string, password: string) => {
      await register(handle, email, password)
      goTo('verify', { email })
    },
    [goTo],
  )

  return { submit: useSubmit(action) }
}
