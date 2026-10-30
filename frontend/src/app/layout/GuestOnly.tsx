/**
 * Pages for someone not signed in: the site, and every screen that signs in
 * or recovers an account.
 */

// --- IMPORTS ---
import { Navigate, Outlet } from 'react-router'

import { SessionPending } from '@/app/layout/SessionPending'
import { useRouteState } from '@/app/navigation/navigation'
import { probablySignedIn, useSession } from '@/logic/session/session'
import { ROUTES } from '@/routes'

// --- CODE ---
/**
 * Renders the page, or moves someone signed in to where they belong.
 *
 * This is the one place that decides where a sign-in lands, so the form
 * only records the session and never navigates on its own.
 *
 * @returns The page, a placeholder, or a redirect.
 */
export function GuestOnly() {
  const session = useSession()
  const state = useRouteState<'login'>()

  // a guest sees the page at once; someone who was signed in last time
  // waits for the answer, so the page never shows before the redirect
  if (session.status === 'loading' && probablySignedIn()) {
    return <SessionPending />
  }

  if (session.status !== 'signed-in') return <Outlet />

  // an account without a handle chooses one before anything else
  if (session.account.handle === null) {
    return <Navigate to={ROUTES.welcome} replace />
  }

  return <Navigate to={state.from ?? ROUTES.dashboard} replace />
}
