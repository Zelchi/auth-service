import { styled } from 'solid-styled-components';
import { Show } from 'solid-js'

const Alert = styled('div') <{ kind: 'error' | 'success' }>`
    position: fixed;
    z-index: 1000;
    top: max(16px, env(safe-area-inset-top));
    left: 50%;
    width: min(420px, calc(100vw - 32px));
    padding: 11px 14px;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    border-radius: 12px;
    font-size: 13px;
    line-height: 1.35;
    border: 1px solid ${p => p.kind === 'error' ? 'var(--error)' : 'var(--success)'};
    color: ${p => p.kind === 'error' ? 'var(--error)' : 'var(--success)'};
    background: ${p => p.kind === 'error' ? 'rgba(255,122,122,0.10)' : 'rgba(114,214,154,0.10)'};
    box-shadow: 0 10px 28px rgba(0, 0, 0, 0.36);
    transform: translateX(-50%);
    animation: toast-in 0.18s ease-out;

    @keyframes toast-in {
        from {
            opacity: 0;
            transform: translate(-50%, -8px);
        }
        to {
            opacity: 1;
            transform: translate(-50%, 0);
        }
    }
`

export default (p: { kind: 'error' | 'success'; message: string }) => (
    <Show when={p.message}>
        <Alert kind={p.kind} role="alert" aria-live={p.kind === 'error' ? 'assertive' : 'polite'}>
            {p.message}
        </Alert>
    </Show>
)
