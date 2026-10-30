/**
 * The platform's chrome, around each app page.
 */

// --- IMPORTS ---
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import { Link, Outlet } from 'react-router'

import { Wordmark } from '@/components/ecoa/Wordmark'
import { Button } from '@/components/ui/Button'
import { Container } from '@/components/ui/Container'
import { cn } from '@/components/ui/cn'
import { useSession } from '@/logic/session/session'
import { ROUTES } from '@/routes'

// --- CODE ---
/**
 * Shows who is signed in, or the way to sign in.
 *
 * @returns The identity slot of the header.
 */
function Identity() {
  const session = useSession()
  const reduceMotion = useReducedMotion()

  // nothing until the session is known, so the slot never flickers
  const content =
    session.status === 'signed-in' ? (
      <Link
        to={session.account.handle ? ROUTES.account : ROUTES.welcome}
        className={cn(
          'font-mono text-mono-m text-secondary transition-colors',
          'duration-interface hover:text-primary',
        )}
      >
        {session.account.handle ? `@${session.account.handle}` : 'Finish setup'}
      </Link>
    ) : session.status === 'signed-out' ? (
      <Button as={Link} to={ROUTES.login} variant="ghost" size="sm">
        Sign in
      </Button>
    ) : null

  return (
    <AnimatePresence mode="wait" initial={false}>
      {content ? (
        <motion.div
          key={session.status}
          initial={reduceMotion ? false : { opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={reduceMotion ? undefined : { opacity: 0 }}
          transition={{ duration: reduceMotion ? 0 : 0.18 }}
        >
          {content}
        </motion.div>
      ) : null}
    </AnimatePresence>
  )
}

/**
 * Wraps every app page in the shared chrome.
 *
 * @returns The chrome, with the matched route rendered inside it.
 */
export function AppLayout() {
  const session = useSession()

  // the site is for visitors: someone signed in is home on the dashboard
  const home = session.status === 'signed-in' ? ROUTES.dashboard : ROUTES.home

  return (
    <>
      <header className="fixed inset-x-0 top-0 z-60 bg-canvas">
        <Container
          className={cn(
            'flex h-[var(--header-h)] items-center justify-between',
          )}
        >
          <Link
            to={home}
            className="flex items-center gap-3"
            aria-label="Ecoa — home"
          >
            <Wordmark />
          </Link>

          <Identity />
        </Container>
      </header>

      <main id="main">
        <Outlet />
      </main>
    </>
  )
}
