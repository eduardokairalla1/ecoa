/**
 * Who is signed in, shared by every screen.
 */

// --- IMPORTS ---
import { useEffect, useSyncExternalStore } from 'react'
import { onSessionLost } from '@/logic/datasource/http'
import { HttpError } from '@/logic/errors/httpError'
import { me } from '@/logic/datasource/auth'
import type { Account } from '@/logic/datasource/auth.t'
import type { Session, SignOutReason } from '@/logic/session/session.t'

// --- GLOBALS ---
let current: Session = { status: 'loading' }
let asked = false
let generation = 0

const listeners = new Set<() => void>()

// whether the last session this browser knew was signed in; only a hint
const HINT_KEY = 'ecoa.signed-in'

// --- CODE ---
/**
 * Remembers, across visits, whether someone was signed in.
 *
 * @param next - The session just published.
 * @returns Nothing.
 */
function remember(next: Session) {
  if (next.status !== 'signed-in' && next.status !== 'signed-out') return

  try {
    localStorage.setItem(HINT_KEY, next.status === 'signed-in' ? '1' : '0')
  } catch {
    // storage blocked: the next visit just goes without the hint
  }
}

/**
 * Guesses, before the backend answers, whether someone is signed in.
 *
 * A hint for what to draw while waiting, never for what to allow.
 *
 * @returns True when the last session this browser knew was signed in.
 */
export function probablySignedIn() {
  try {
    return localStorage.getItem(HINT_KEY) === '1'
  } catch {
    return false
  }
}

/**
 * Replaces the session and tells every subscriber.
 *
 * @param next - The new session.
 * @returns Nothing.
 */
function publish(next: Session) {
  current = next
  remember(next)
  listeners.forEach((listener) => listener())
}

/**
 * Asks the backend who is signed in.
 *
 * @returns Nothing.
 */
export async function loadSession() {

  // mark that the session has been asked for
  asked = true

  // generate a unique identifier for this request
  const mine = ++generation

  // publish the loading state
  publish({ status: 'loading' })

  // ask the backend who is signed in
  let next: Session
  try {
    next = { status: 'signed-in', account: await me() }

  // nobody is signed in: publish that
  } catch (error) {
    const signedOut = error instanceof HttpError && error.status === 401
    next = { status: signedOut ? 'signed-out' : 'unavailable' }
  }

  // a sign-in or sign-out happened meanwhile: it knows better
  if (mine === generation) publish(next)
}

/**
 * Records the account that just signed in, or changed.
 *
 * @param account - The account.
 * @returns Nothing.
 */
export function signedIn(account: Account) {
  asked = true
  generation++
  publish({ status: 'signed-in', account })
}

/**
 * Records that nobody is signed in any more.
 *
 * @param reason - What ended the session, when the person did it here.
 * @returns Nothing.
 */
export function signedOut(reason?: SignOutReason) {
  asked = true
  generation++
  publish({ status: 'signed-out', reason })
}

/**
 * Forgets why the session ended, once the sign-in screen has said so.
 *
 * @returns Nothing.
 */
export function settleSignOut() {
  if (current.status === 'signed-out' && current.reason) {
    publish({ status: 'signed-out' })
  }
}

// a session that could not be renewed ends here too
onSessionLost(() => {
  if (current.status === 'signed-in') signedOut()
})

/**
 * Subscribes to the session.
 *
 * @param listener - Called on every change.
 * @returns The function that unsubscribes.
 */
function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

/**
 * Reads the session, asking the backend the first time it is needed.
 *
 * @returns The current session.
 */
export function useSession() {
  const session = useSyncExternalStore(subscribe, () => current)

  useEffect(() => {
    if (!asked) void loadSession()
  }, [])

  return session
}
