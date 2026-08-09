import {
    RESPONSE_TYPE,
    extractToken,
    isAllowedOrigin,
    isBridgeRequest,
    parseAllowedOrigins,
} from './bridgeProtocol.mjs'

const allowedOrigins = parseAllowedOrigins(import.meta.env.VITE_AUTH_BRIDGE_ORIGINS ?? '')

const requestSessionToken = async (): Promise<string | null> => {
    try {
        const response = await fetch('/api/bridge/token', {
            method: 'POST',
            credentials: 'same-origin',
            headers: { 'Content-Type': 'application/json' },
        })
        if (!response.ok) return null

        const data: unknown = await response.json()
        if (!data || typeof data !== 'object') return null

        return extractToken(data)
    } catch {
        return null
    }
}

export const installBridge = (origins = allowedOrigins) => {
    const handleMessage = async (event: MessageEvent) => {
        if (event.source !== window.parent) return
        if (!isAllowedOrigin(event.origin, origins)) return
        if (!isBridgeRequest(event.data)) return

        const token = await requestSessionToken()

        event.source.postMessage({
            type: RESPONSE_TYPE,
            requestId: event.data.requestId,
            token,
        }, {
            targetOrigin: event.origin,
        })
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
}
