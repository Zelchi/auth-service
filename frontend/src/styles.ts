import { styled } from 'solid-styled-components'

export const Button = styled('button') <{ loading?: boolean }>`
    width: 100%;
    min-height: 42px;
    margin-top: 0;
    padding: 0 16px;
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
    transition: background 0.15s, opacity 0.15s, transform 0.15s;
    opacity: ${p => p.loading ? '0.6' : '1'};
    pointer-events: ${p => p.loading ? 'none' : 'auto'};
    &:hover {
        background: var(--accent-hover);
        transform: translateY(-1px);
    }
    &:active { transform: translateY(0); }
`

export const Card = styled('div')`
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 16px;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.32);
    padding: 28px 24px;
    width: 100%;
    max-width: 380px;
    display: flex;
    flex-direction: column;
    gap: 16px;

    @media (max-width: 480px) {
        border-radius: 14px;
        padding: 24px 20px;
    }
`

export const Center = styled('div')`
    background: var(--bg);
    min-height: 100dvh;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 64px 20px 24px;
    overflow-y: auto;
`

export const Title = styled('h1')`
    font-size: 21px;
    font-weight: 700;
    color: var(--text);
    letter-spacing: -0.02em;
    margin: 0;
`

export const Subtitle = styled('p')`
    font-size: 13px;
    color: var(--muted);
    margin: 4px 0 0;
    line-height: 1.45;
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
