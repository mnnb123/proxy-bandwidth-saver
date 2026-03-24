import { useEffect, useRef } from 'react'

const isWails = typeof (window as any).__wails_invoke !== 'undefined'

// Singleton SSE connection for web mode
let sseSource: EventSource | null = null
const sseListeners = new Map<string, Set<(data: any) => void>>()

function getSSE(): EventSource {
  if (!sseSource || sseSource.readyState === EventSource.CLOSED) {
    sseSource = new EventSource('/api/events')
    sseSource.onerror = () => {
      setTimeout(() => {
        if (sseSource?.readyState === EventSource.CLOSED) {
          sseSource = null
        }
      }, 3000)
    }
  }
  return sseSource
}

function addSSEListener(event: string, cb: (data: any) => void) {
  if (!sseListeners.has(event)) {
    sseListeners.set(event, new Set())
    const sse = getSSE()
    sse.addEventListener(event, (e: MessageEvent) => {
      try {
        const data = JSON.parse(e.data)
        const cbs = sseListeners.get(event)
        if (cbs) cbs.forEach((fn) => fn(data))
      } catch { /* ignore */ }
    })
  }
  sseListeners.get(event)!.add(cb)
}

function removeSSEListener(event: string, cb: (data: any) => void) {
  const cbs = sseListeners.get(event)
  if (cbs) cbs.delete(cb)
}

export function useWailsEvent<T = unknown>(eventName: string, handler: (data: T) => void) {
  const handlerRef = useRef(handler)
  handlerRef.current = handler

  useEffect(() => {
    const cb = (data: T) => handlerRef.current(data)

    if (isWails) {
      let cancel: (() => void) | undefined
      import('../../wailsjs/runtime/runtime').then(({ EventsOn }) => {
        cancel = EventsOn(eventName, cb)
      })
      return () => { cancel?.() }
    } else {
      addSSEListener(eventName, cb)
      return () => removeSSEListener(eventName, cb)
    }
  }, [eventName])
}
