import React from 'react'
import { Button, Tooltip, Result, Space } from 'antd'
import { LoginOutlined } from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

interface Props {
  children: React.ReactNode
  /** Action description shown to unauthenticated users, e.g. "comment on this article" */
  action?: string
  /** Render mode:
   *   'tooltip' — wrap children, show tooltip on interaction attempt for inline buttons
   *   'block'   — replace children entirely with a "please login" prompt
   *   'hidden'  — don't render children at all for unauthenticated users
   */
  mode?: 'tooltip' | 'block' | 'hidden'
  /** Fallback element when blocked (only for 'block' mode) */
  fallback?: React.ReactNode
}

export default function PermissionGuard({
  children,
  action = 'perform this action',
  mode = 'tooltip',
  fallback,
}: Props) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const navigate = useNavigate()
  const location = useLocation()

  if (isAuthenticated) {
    return <>{children}</>
  }

  const loginUrl = `/login?redirect=${encodeURIComponent(location.pathname + location.search)}`

  // ============ Hidden mode ============
  if (mode === 'hidden') {
    return null
  }

  // ============ Block mode ============
  if (mode === 'block') {
    if (fallback) return <>{fallback}</>
    return (
      <Result
        status="403"
        title="Login Required"
        subTitle={`Please sign in to ${action}`}
        extra={
          <Button type="primary" icon={<LoginOutlined />} onClick={() => navigate(loginUrl)}>
            Sign In Now
          </Button>
        }
      />
    )
  }

  // ============ Tooltip mode (default) ============
  return (
    <Tooltip title={`Sign in to ${action}`} placement="top">
      <span
        style={{ cursor: 'not-allowed', display: 'inline-flex' }}
        onClick={(e) => {
          e.preventDefault()
          e.stopPropagation()
          navigate(loginUrl)
        }}
      >
        {/* Clone children but force disabled state */}
        {React.Children.map(children, (child) => {
          if (React.isValidElement(child)) {
            return React.cloneElement(child as React.ReactElement<any>, {
              disabled: true,
              onClick: (e: React.MouseEvent) => {
                e.preventDefault()
                e.stopPropagation()
                navigate(loginUrl)
              },
            })
          }
          return child
        })}
      </span>
    </Tooltip>
  )
}
