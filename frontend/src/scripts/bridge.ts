import Cookies from 'js-cookie'

export const installBridge = () => {
    const handleMessage = (event: MessageEvent) => {

        if (event.source !== window.parent) return

        event.source.postMessage({
            type: 'AUTH_SERVICE/TOKEN_RESPONSE',
            requestId: event.data.requestId,
            token: Cookies.get('auth-token') ?? null,
        }, {
            targetOrigin: event.origin,
        })
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
}