/**
 * Every destination the frontend links to.
 */

// --- GLOBALS ---
// the platform lives under /app
const APP = '/app'
const PENDING = '/404'

export const ROUTES = {
  home: '/',
  dashboard: APP,
  explore: PENDING,
  write: PENDING,
  login: `${APP}/login`,
  register: `${APP}/register`,
  verify: `${APP}/verify`,
  forgotPassword: `${APP}/forgot-password`,
  resetPassword: `${APP}/reset-password`,
  welcome: `${APP}/welcome`,
} as const
