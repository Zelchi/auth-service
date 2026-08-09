export declare const REQUEST_TYPE: string
export declare const RESPONSE_TYPE: string
export declare const MAX_REQUEST_ID_LENGTH: number

export declare const parseAllowedOrigins: (value?: string) => Set<string>
export declare const isAllowedOrigin: (origin: string, allowedOrigins: Set<string>) => boolean
export declare const isBridgeRequest: (data: unknown) => data is {
    type: string
    requestId: string
}
export declare const extractToken: (data: unknown) => string | null
