const BASE = '/api'

export interface AuthenticatedUser {
    id: string
    email: string
    name: string
    image: string
    created_at: string
}

const request = async <T>(path: string, options?: RequestInit): Promise<T> => {
    const controller = new AbortController()
    const timeout = setTimeout(() => controller.abort(), 15000)
    let res: Response

    try {
        res = await fetch(`${BASE}${path}`, {
            credentials: 'same-origin',
            headers: { 'Content-Type': 'application/json' },
            ...options,
            signal: controller.signal,
        })
    } catch (error: unknown) {
        if (error instanceof Error && error.name === 'AbortError') {
            throw new Error('O serviço de autenticação demorou para responder.')
        }
        throw error
    } finally {
        clearTimeout(timeout)
    }

    const raw = await res.text()
    let data: unknown = null
    if (raw) {
        try {
            data = JSON.parse(raw)
        } catch {
            data = null
        }
    }

    if (!res.ok) {
        const message = data && typeof data === 'object' && 'error' in data
            && typeof data.error === 'string'
            ? data.error
            : 'Erro desconhecido'
        throw new Error(message)
    }

    return data as T
}

export default {
    register: (email: string, password: string, passwordConfirmation: string) => request<{ message: string }>('/register', {
        method: 'POST',
        body: JSON.stringify({ email, password, password_confirmation: passwordConfirmation }),
    }),

    verify: (email: string, code: string) => request<{ message: string }>('/verify', {
        method: 'POST',
        body: JSON.stringify({ email, code }),
    }),

    resend: (email: string) => request<{ message: string }>('/resend', {
        method: 'POST',
        body: JSON.stringify({ email }),
    }),

    login: (email: string, password: string) => request<{ user_id: string }>('/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    }),

    logout: () => request<{ message: string }>('/logout', {
        method: 'POST',
    }),

    me: () => request<AuthenticatedUser>('/me'),

    updateProfile: (name: string, image: string) => request<AuthenticatedUser>('/me', {
        method: 'PATCH',
        body: JSON.stringify({ name, image }),
    }),
}
