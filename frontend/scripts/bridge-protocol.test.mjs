import assert from 'node:assert/strict'
import test from 'node:test'

import {
    REQUEST_TYPE,
    extractToken,
    isAllowedOrigin,
    isBridgeRequest,
    parseAllowedOrigins,
} from '../src/scripts/bridgeProtocol.mjs'

test('accepts only the bridge request contract', () => {
    assert.equal(isBridgeRequest({ type: REQUEST_TYPE, requestId: 'request-123' }), true)
    assert.equal(isBridgeRequest({ type: 'OTHER', requestId: 'request-123' }), false)
    assert.equal(isBridgeRequest({ type: REQUEST_TYPE }), false)
    assert.equal(isBridgeRequest({ type: REQUEST_TYPE, requestId: 'contains space' }), false)
})

test('matches only configured origins', () => {
    const origins = parseAllowedOrigins('https://trusted.example, https://second.example')
    assert.equal(isAllowedOrigin('https://trusted.example', origins), true)
    assert.equal(isAllowedOrigin('https://evil.example', origins), false)
})

test('does not treat a missing or malformed token as a session token', () => {
    assert.equal(extractToken({}), null)
    assert.equal(extractToken({ token: 123 }), null)
    assert.equal(extractToken({ token: '' }), null)
    assert.equal(extractToken({ token: 'jwt-value' }), 'jwt-value')
})
