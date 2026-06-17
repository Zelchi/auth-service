import { styled } from 'solid-styled-components';
import { Show } from 'solid-js'

const Alert = styled('div') <{ kind: 'error' | 'success' }>`
    position: absolute;
    top: 35px;
    right: 24px;
    padding: 11px 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    font-size: 14px;
    border: 1px solid ${p => p.kind === 'error' ? 'var(--error)' : 'var(--success)'};
    color: ${p => p.kind === 'error' ? 'var(--error)' : 'var(--success)'};
    background: ${p => p.kind === 'error' ? 'rgba(192,57,43,0.08)' : 'rgba(46,204,113,0.08)'};
`

export default (p: { kind: 'error' | 'success'; message: string }) => (
    <Show when={p.message}>
        <Alert kind={p.kind}>{p.message}</Alert>
    </Show>
)