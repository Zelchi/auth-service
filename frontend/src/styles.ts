import { styled } from 'solid-styled-components'

export const Button = styled('button') <{ loading?: boolean }>`
    width: 100%;
    height: 40px;
    margin-top: 12px;
    padding-bottom: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--accent);
    color: var(--bg);
    border: none;
    border-radius: 8px;
    font-family: inherit;
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 0.02em;
    cursor: pointer;
    transition: background 0.15s, opacity 0.15s;
    opacity: ${p => p.loading ? '0.6' : '1'};
    pointer-events: ${p => p.loading ? 'none' : 'auto'};
    &:hover {
        background: var(--accent-hover);
    }
    @media (prefers-color-scheme: light) {
        color: #fff;
    }
`

export const Card = styled('div')`
    position: relative;
    background: var(--surface);
    padding: 36px 32px;
    width: 100%;
    max-width: 400px;
    display: flex;
    flex-direction: column;
    gap: 20px;
`

export const Center = styled('div')`
    background: var(--surface);
    height: 100dvh;
    width: 100dvw;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
`

export const Title = styled('h1')`
    font-size: 22px;
    font-weight: 700;
    color: var(--text);
    letter-spacing: -0.02em;
    margin: 0;
`

export const Subtitle = styled('p')`
    font-size: 14px;
    color: var(--muted);
    margin: -8px 0 0;
`

export const TextLink = styled('button')`
    background: none;
    border: none;
    color: var(--accent);
    font-family: inherit;
    font-size: 14px;
    cursor: pointer;
    padding: 0;
    text-decoration: underline;
    text-underline-offset: 3px;
    &:hover { color: var(--accent-hover); }
`