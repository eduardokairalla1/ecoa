/**
 * The route tree for the site and the platform.
 */

// --- IMPORTS ---
import { createBrowserRouter } from 'react-router'

import { Register } from '@/app/auth/register/Register'
import { VerifyEmail } from '@/app/auth/verifyEmail/VerifyEmail'
import { AppLayout } from '@/app/layout/AppLayout'
import { ForgotPassword } from '@/app/auth/ForgotPassword'
import { Login } from '@/app/auth/Login'
import { ResetPassword } from '@/app/auth/ResetPassword'
import { Landing } from '@/site/landing/Landing'
import { NotFound } from '@/site/NotFound'
import { SiteLayout } from '@/site/layout/SiteLayout'

// --- GLOBALS ---
export const router = createBrowserRouter([
  {
    element: <SiteLayout />,
    children: [{ index: true, element: <Landing /> }],
  },
  {
    path: 'app',
    element: <AppLayout />,
    children: [
      { path: 'login', element: <Login /> },
      { path: 'register', element: <Register /> },
      { path: 'verify', element: <VerifyEmail /> },
      { path: 'forgot-password', element: <ForgotPassword /> },
      { path: 'reset-password', element: <ResetPassword /> },
    ],
  },
  {
    element: <SiteLayout nav={false} />,
    children: [{ path: '*', element: <NotFound /> }],
  },
])
