/**
 * Moving between screens, with what one hands the next typed on both ends.
 */

// --- IMPORTS ---
import { useCallback } from 'react'
import { useLocation, useNavigate } from 'react-router'

import type { RouteStates } from '@/app/navigation/navigation.t'
import { ROUTES } from '@/routes'

// --- CODE ---
/**
 * Builds the props of a Link or a Navigate to a screen that takes a state.
 *
 * @param route - The screen, by its ROUTES name.
 * @param state - What it is handed.
 * @returns The `to` and `state` props.
 */
export function routeTo<R extends keyof RouteStates>(
  route: R,
  state: RouteStates[R],
) {
  return { to: ROUTES[route], state }
}

/**
 * Navigates to a screen that takes a state, checking what is handed over.
 *
 * @returns The function that navigates.
 */
export function useGoTo() {
  const navigate = useNavigate()

  return useCallback(
    <R extends keyof RouteStates>(route: R, state: RouteStates[R]) => {
      void navigate(ROUTES[route], { state })
    },
    [navigate],
  )
}

/**
 * Reads what the previous screen handed this one, empty on a direct visit.
 *
 * @returns The state, typed by the screen it is read on.
 */
export function useRouteState<R extends keyof RouteStates>() {
  return (useLocation().state ?? {}) as RouteStates[R]
}
