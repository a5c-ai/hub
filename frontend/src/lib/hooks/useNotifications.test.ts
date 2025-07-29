import { renderHook, act } from '@testing-library/react'
import { useNotifications } from './useNotifications'

// Mock WebSocket
class MockWebSocket {
  onmessage: ((event: { data: string }) => void) | null = null
  onerror: (() => void) | null = null
  close = jest.fn()
  send(data: string) {
    // simulate server message
    if (this.onmessage) this.onmessage({ data })
  }
}

describe('useNotifications', () => {
  let originalWebSocket: any

  beforeAll(() => {
    originalWebSocket = (global as any).WebSocket
    ;(global as any).WebSocket = MockWebSocket
    process.env.NEXT_PUBLIC_API_URL = 'http://localhost/api/v1'
  })
  afterAll(() => {
    (global as any).WebSocket = originalWebSocket
  })

  it('receives notifications via WebSocket', () => {
    const { result } = renderHook(() => useNotifications<any>())
    const ws = (result.all[result.all.length - 1] as any).current
    // send test notification
    act(() => {
      ;(ws as MockWebSocket).send(JSON.stringify({ id: '1', message: 'hi' }))
    })
    expect(result.current).toEqual([{ id: '1', message: 'hi' }])
  })
})
