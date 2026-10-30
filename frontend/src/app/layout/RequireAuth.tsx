/**
 * The guarded part of the platform: screens that need someone signed in.
 */

// --- IMPORTS ---
import { Navigate, Outlet, useLocation } from 'react-router'

import type { RequireAuthProps } from '@/app/layout/RequireAuth.t'
import { routeTo } from '@/app/navigation/navigation'
import { SessionPending } from '@/app/layout/SessionPending'
import { SessionUnavailable } from '@/app/layout/SessionUnavailable'
import { useSession } from '@/logic/session/session'
import { ROUTES } from '@/routes'

// --- CODE ---
/**
 * Renders the guarded screen, or sends the person where they belong.
 *
 * @param props - Whether the screens need a complete profile.
 * @returns The screen, a placeholder, or a redirect.
 */
export function RequireAuth({ profile }: RequireAuthProps) {
  const session = useSession()
  const location = useLocation()

  if (session.status === 'loading') return <SessionPending />
  if (session.status === 'unavailable') return <SessionUnavailable />

  // ended here on purpose: say so on the way out, and do not come back
  if (session.status === 'signed-out' && session.reason) {
    return <Navigate {...routeTo('login', { ended: session.reason })} replace />
  }

  // never signed in: sign in, then come back here
  if (session.status === 'signed-out') {
    const from = location.pathname + location.search + location.hash
    return <Navigate {...routeTo('login', { from })} replace />
  }

  const complete = session.account.handle !== null

  // a screen that needs a handle, for an account without one
  if (profile && !complete) return <Navigate to={ROUTES.welcome} replace />

  // the handle screen, for an account that already chose it
  if (!profile && complete) return <Navigate to={ROUTES.dashboard} replace />

  return <Outlet />
}
