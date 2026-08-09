export declare const PENDING_EMAIL_KEY: string
export declare const PENDING_TTL: number
export declare const clearPendingEmail: (storage: Storage) => void
export declare const readPendingEmail: (storage: Storage, now?: number) => string
export declare const writePendingEmail: (storage: Storage, email: string, now?: number) => boolean
