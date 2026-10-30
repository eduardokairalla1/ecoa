/**
 * Types for the Button.
 */

// --- IMPORTS ---
import type { ComponentPropsWithoutRef, ElementType, ReactNode } from 'react'

// --- GLOBALS ---
export type ButtonVariant = 'primary' | 'secondary' | 'ghost'

export type ButtonSize = 'sm' | 'md' | 'lg'

export type ButtonProps<T extends ElementType> = {
  as?: T
  variant?: ButtonVariant
  size?: ButtonSize
  /** Waiting on the action it started: disabled, with a running mark. */
  pending?: boolean
  children: ReactNode
} & Omit<ComponentPropsWithoutRef<T>, 'as' | 'children'>
