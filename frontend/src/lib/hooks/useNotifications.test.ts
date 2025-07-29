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
  let wsInstance: MockWebSocket

  beforeAll(() => {
    originalWebSocket = (global as any).WebSocket
    ;(global as any).WebSocket = class extends MockWebSocket {
      constructor(url: string) {
        super()
        // eslint-disable-next-line @typescript-eslint/no-this-alias
        wsInstance = this
      }
    } as any
    process.env.NEXT_PUBLIC_API_URL = 'http://localhost/api/v1'
  })
  afterAll(() => {
    (global as any).WebSocket = originalWebSocket
  })

  it('receives notifications via WebSocket', () => {
    const { result } = renderHook(() => useNotifications<any>())
    // send test notification
    act(() => {
      wsInstance.send(JSON.stringify({ id: '1', message: 'hi' }))
    })
    expect(result.current).toEqual([{ id: '1', message: 'hi' }])
  })
})
