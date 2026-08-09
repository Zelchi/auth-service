const BASE = '/api'

const request = async <T>(path: string, options?: RequestInit): Promise<T> => {
    const res = await fetch(`${BASE}${path}`, {
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/json' },
        ...options,
    })

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
    register: (email: string, password: string) => request<{ message: string }>('/register', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
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

    me: () => request<{ id: string; email: string; created_at: string }>('/me'),
}
