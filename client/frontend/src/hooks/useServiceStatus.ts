import { useState, useEffect } from 'react'
import type { ServiceStatus } from '../lib/types'

// Wails v3 event system - will use runtime.EventsOn once bindings are generated
export function useServiceStatus(initial: ServiceStatus = 'idle') {
  const [status] = useState<ServiceStatus>(initial)

  useEffect(() => {
    // This will be connected to Wails events once bindings are generated
    // For now, use a simple polling approach or direct call
    const interval = setInterval(async () => {
      try {
        // Will call GetServiceStatus from Wails bindings
        // const newStatus = await GetServiceStatus()
        // setStatus(newStatus)
      } catch {
        // ignore errors during polling
      }
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  return status
}
