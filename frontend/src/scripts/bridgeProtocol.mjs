export const REQUEST_TYPE = 'AUTH_SERVICE/TOKEN_REQUEST'
export const RESPONSE_TYPE = 'AUTH_SERVICE/TOKEN_RESPONSE'
export const MAX_REQUEST_ID_LENGTH = 128

export const parseAllowedOrigins = (value = '') => new Set(
    value
        .split(',')
        .map(origin => origin.trim())
        .filter(Boolean),
)

export const isAllowedOrigin = (origin, allowedOrigins) => allowedOrigins.has(origin)

export const isBridgeRequest = (data) => {
    if (!data || typeof data !== 'object') return false

    const message = data
    return message.type === REQUEST_TYPE
        && typeof message.requestId === 'string'
        && /^[A-Za-z0-9._:-]{1,128}$/.test(message.requestId)
        && message.requestId.length <= MAX_REQUEST_ID_LENGTH
}

export const extractToken = (data) => {
    if (!data || typeof data !== 'object') return null

    const token = data.token
    return typeof token === 'string' && token.length > 0 ? token : null
}
