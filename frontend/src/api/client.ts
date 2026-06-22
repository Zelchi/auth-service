import Cookies from 'js-cookie'

const BASE = '/api'

const request = async <T>(path: string, options?: RequestInit): Promise<T> => {
    const token = Cookies.get('auth-token')

    const res = await fetch(`${BASE}${path}`, {
        headers: {
            'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        ...options,
    })

    const data = await res.json()

    if (!res.ok) {
        throw new Error(data.error || 'Erro desconhecido')
    }

    return data as T
}

export default {
    register: (email: string, password: string) => request<{ user_id: string; message: string }>('/register', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    }),

    verify: (email: string, code: string) => request<{ message: string }>('/verify', {
        method: 'POST',
        body: JSON.stringify({ email, code }),
    }),

    login: (email: string, password: string) => request<{ token: string; user_id: string }>('/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    }),

    me: () => request<{ id: string; email: string; created_at: string }>('/me'),
}