/**
 * A wait between two requests for a new code.
 */

// --- IMPORTS ---
import { useCallback, useEffect, useState } from 'react'

// --- CODE ---
/**
 * Counts down after each start, so a code is not asked for twice in a row.
 *
 * @param seconds - How long each wait lasts.
 * @param running - Whether a wait is already under way on arrival.
 * @returns The seconds left, 0 when ready, and the function that starts it.
 */
export function useCooldown(seconds: number, running = false) {
  const [left, setLeft] = useState(running ? seconds : 0)

  useEffect(() => {
    if (left <= 0) return

    const timer = setTimeout(() => setLeft((value) => value - 1), 1000)
    return () => clearTimeout(timer)
  }, [left])

  const start = useCallback(() => setLeft(seconds), [seconds])

  return { left, start }
}
