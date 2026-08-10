const themes = {
    zelchi: {
        bg: '#0a0a0a',
        surface: '#171717',
        border: '#2a2a2a',
        text: '#ededed',
        muted: '#888888',
        accent: '#ededed',
        accentHover: '#ffffff',
        error: '#f19a9a',
        success: '#9bd5a8',
    },
} as const

type ThemeName = keyof typeof themes

const defaultTheme: ThemeName = 'zelchi'

function resolveTheme(value: string | null | undefined): ThemeName {
    const candidate = value?.trim().toLowerCase() ?? ''
    return Object.prototype.hasOwnProperty.call(themes, candidate)
        ? candidate as ThemeName
        : defaultTheme
}

export function applyAuthTheme(value = new URLSearchParams(window.location.search).get('theme')) {
    const name = resolveTheme(value)
    const theme = themes[name]
    const root = document.documentElement

    for (const [key, color] of Object.entries(theme)) {
        root.style.setProperty(`--${key.replace(/[A-Z]/g, match => `-${match.toLowerCase()}`)}`, color)
    }

    root.dataset.authTheme = name
}
