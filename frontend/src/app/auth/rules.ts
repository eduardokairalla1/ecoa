/**
 * What the backend accepts, checked before asking it.
 */

// --- GLOBALS ---
// lowercase letters, digits and '_', starting with a letter, 3 to 30 long;
// the same rule the backend and the database hold
export const HANDLE = /^[a-z][a-z0-9_]{2,29}$/

export const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export const CODE_LENGTH = 6

export const MIN_PASSWORD = 8

// seconds before another code can be asked for
export const RESEND_WAIT = 30

// the symbols the identity provider counts, space included
const SYMBOL = /[\^$*.[\]{}()?"!@#%&/\\,><':;|_~`=+\- ]/

/** The identity provider's password policy, one rule per line. */
export const PASSWORD_RULES = [
  {
    label: `${MIN_PASSWORD} characters or more`,
    test: new RegExp(`^.{${MIN_PASSWORD},}$`),
  },
  { label: 'A lowercase letter', test: /[a-z]/ },
  { label: 'An uppercase letter', test: /[A-Z]/ },
  { label: 'A number', test: /[0-9]/ },
  { label: 'A symbol', test: SYMBOL },
] as const

/** What each field says when what was typed breaks its rule. */
export const INVALID = {
  email: 'Enter a valid email address.',
  code: `Enter all ${CODE_LENGTH} digits.`,
  handle:
    'Lowercase letters, numbers and _, starting with a letter. 3 to 30 long.',
  password: 'Meet every rule below.',
  confirm: 'The two do not match.',
} as const

// --- CODE ---
/**
 * Tells whether a password meets the whole policy.
 *
 * @param password - The password typed.
 * @returns True when every rule holds.
 */
export function passwordMeetsPolicy(password: string) {
  return PASSWORD_RULES.every((rule) => rule.test.test(password))
}
