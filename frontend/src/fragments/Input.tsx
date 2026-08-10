import { Show, splitProps, type JSX } from 'solid-js'
import { styled } from 'solid-styled-components'

const InputEl = styled('input')`
    width: 100%;
    min-height: 42px;
    padding: 10px 12px;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 10px;
    color: var(--text);
    font-family: inherit;
    font-size: 15px;
    transition: border-color 0.15s, box-shadow 0.15s;

    &:focus {
        border-color: var(--accent);
        box-shadow: 0 0 0 3px rgba(236, 236, 236, 0.14);
    }

    &::placeholder {
        color: var(--muted);
    }
`

type InputProps = {
    label: string
} & JSX.InputHTMLAttributes<HTMLInputElement>

export default (props: InputProps) => {
    const [local, inputProps] = splitProps(props, ['label'])
    return (
        <div style={{ display: 'flex', 'flex-direction': 'column', gap: '6px', width: '100%' }}>
            <Show when={local.label}>
                <label style={{ 'font-size': '12px', color: 'var(--muted)', 'letter-spacing': '0.04em' }}>
                    {local.label}
                </label>
            </Show>
            <InputEl {...inputProps} />
        </div>
    )
}
