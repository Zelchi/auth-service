import assert from 'node:assert/strict'
import test from 'node:test'

import {
    PENDING_EMAIL_KEY,
    PENDING_TTL,
    readPendingEmail,
    writePendingEmail,
} from '../src/pendingStorage.mjs'

const makeStorage = () => {
    const values = new Map()
    return {
        getItem: key => values.get(key) ?? null,
        setItem: (key, value) => values.set(key, value),
        removeItem: key => values.delete(key),
    }
}

test('persists and normalizes a valid pending email', () => {
    const storage = makeStorage()
    assert.equal(writePendingEmail(storage, ' USER@Example.COM ', 1000), true)
    assert.equal(readPendingEmail(storage, 1000 + 1000), 'user@example.com')
})

test('clears expired or malformed pending values', () => {
    const storage = makeStorage()
    storage.setItem(PENDING_EMAIL_KEY, JSON.stringify({ createdAt: 1000, email: 'user@example.com' }))
    assert.equal(readPendingEmail(storage, 1000 + PENDING_TTL + 1), '')
    assert.equal(storage.getItem(PENDING_EMAIL_KEY), null)

    storage.setItem(PENDING_EMAIL_KEY, 'not-valid')
    assert.equal(readPendingEmail(storage, 1000), '')
    assert.equal(storage.getItem(PENDING_EMAIL_KEY), null)
})
