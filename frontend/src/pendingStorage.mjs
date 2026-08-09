export const PENDING_EMAIL_KEY = 'pending-email'
export const PENDING_TTL = 15 * 60 * 1000

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export const clearPendingEmail = (storage) => {
    try {
        storage.removeItem(PENDING_EMAIL_KEY)
    } catch {
        // O armazenamento pode estar indisponível no modo privado do browser.
    }
}

export const readPendingEmail = (storage, now = Date.now()) => {
    try {
        const raw = storage.getItem(PENDING_EMAIL_KEY)
        if (!raw) return ''

        let createdAt = 0
        let email = ''

        if (raw.startsWith('{')) {
            const parsed = JSON.parse(raw)
            if (parsed && typeof parsed === 'object') {
                createdAt = typeof parsed.createdAt === 'number' ? parsed.createdAt : 0
                email = typeof parsed.email === 'string' ? parsed.email : ''
            }
        } else {
            const separator = raw.indexOf('|')
            if (separator > 0) {
                createdAt = Number(raw.slice(0, separator))
                email = raw.slice(separator + 1)
            }
        }

        const isValidTimestamp = Number.isSafeInteger(createdAt)
            && createdAt > 0
            && createdAt <= now + 60_000
            && now - createdAt <= PENDING_TTL
        const normalizedEmail = email.trim().toLowerCase()

        if (!isValidTimestamp || normalizedEmail.length > 254 || !emailPattern.test(normalizedEmail)) {
            clearPendingEmail(storage)
            return ''
        }

        return normalizedEmail
    } catch {
        clearPendingEmail(storage)
        return ''
    }
}

export const writePendingEmail = (storage, email, now = Date.now()) => {
    const normalizedEmail = email.trim().toLowerCase()
    if (normalizedEmail.length > 254 || !emailPattern.test(normalizedEmail)) {
        clearPendingEmail(storage)
        return false
    }

    try {
        storage.setItem(PENDING_EMAIL_KEY, JSON.stringify({
            createdAt: now,
            email: normalizedEmail,
        }))
        return true
    } catch {
        return false
    }
}
