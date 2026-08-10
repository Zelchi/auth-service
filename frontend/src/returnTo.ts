import { parseAllowedOrigins } from './scripts/bridgeProtocol.mjs'

export const AUTHENTICATED_MESSAGE = 'AUTH_SERVICE/AUTHENTICATED'

const allowedReturnOrigins = parseAllowedOrigins(import.meta.env.VITE_AUTH_RETURN_ORIGINS ?? '')

export function safeReturnTo(value: string | null | undefined) {
    if (!value) return '/dashboard'

    try {
        const candidate = new URL(value, window.location.origin)
        const validProtocol = candidate.protocol === 'http:' || candidate.protocol === 'https:'
        if (!validProtocol || candidate.username || candidate.password) return '/dashboard'

        if (candidate.origin === window.location.origin) {
            return `${candidate.pathname}${candidate.search}${candidate.hash}`
        }
        if (allowedReturnOrigins.has(candidate.origin)) return candidate.toString()
    } catch {
        // A URL inválida não pode controlar o redirecionamento após o login.
    }

    return '/dashboard'
}

export function completeReturnTo(target: string) {
    const destination = new URL(target, window.location.origin)

    if (window.parent !== window) {
        window.parent.postMessage({
            type: AUTHENTICATED_MESSAGE,
            returnTo: destination.toString(),
        }, destination.origin)
        return
    }

    window.location.assign(destination.toString())
}
