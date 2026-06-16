import { useState, useEffect } from 'react'
import { App } from '../../bindings/github.com/chenhg5/cc-connect/client'
import { Events } from '@wailsio/runtime'
import type { ServiceStatus } from '../lib/types'

interface ServiceStatusResult {
  status: ServiceStatus
  errorMessage: string
}

export function useServiceStatus(initial: ServiceStatus = 'idle'): ServiceStatusResult {
  const [status, setStatus] = useState<ServiceStatus>(initial)
  const [errorMessage, setErrorMessage] = useState('')

  useEffect(() => {
    // Poll status on mount
    App.GetServiceStatus().then((s: string) => {
      setStatus(s as ServiceStatus)
    }).catch(() => {
      // ignore errors
    })

    // Listen for Wails events to update status in real-time
    const unsubs: (() => void)[] = []

    unsubs.push(Events.On('service:starting', () => {
      setStatus('starting')
      setErrorMessage('')
    }))
    unsubs.push(Events.On('service:running', () => {
      setStatus('running')
      setErrorMessage('')
    }))
    unsubs.push(Events.On('service:stopping', () => {
      setStatus('stopping')
      setErrorMessage('')
    }))
    unsubs.push(Events.On('service:idle', () => {
      setStatus('idle')
      setErrorMessage('')
    }))
    unsubs.push(Events.On('service:error', (data: any) => {
      setStatus('error')
      // The event data contains the error message string from service.go
      setErrorMessage(data?.data || data || '')
    }))

    return () => {
      for (const unsub of unsubs) {
        unsub()
      }
    }
  }, [])

  return { status, errorMessage }
}
