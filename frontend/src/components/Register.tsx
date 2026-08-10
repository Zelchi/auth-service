import { createEffect, createMemo, createSignal, For, Show } from 'solid-js'
import { Button, Card, Center, Title, TextLink } from '../styles'
import Input from '../fragments/Input'
import Alert from '../fragments/Alert'
import { getPasswordStrength } from '../passwordStrength'

import API from '../api/client'
import { errorMessage } from '../api/errorMessage'

interface Props {
    onRegistered: (email: string) => void
    onLogin: () => void
}

export default (props: Props) => {
    const [email, setEmail] = createSignal('')
    const [password, setPassword] = createSignal('')
    const [passwordConfirmation, setPasswordConfirmation] = createSignal('')
    const [loading, setLoading] = createSignal(false)
    const [error, setError] = createSignal('')
    const strength = createMemo(() => getPasswordStrength(password()))
    const passwordsMatch = createMemo(() => passwordConfirmation().length > 0 && password() === passwordConfirmation())

    const submit = async () => {
        if (loading()) return

        setError('')
        if (!strength().isStrong) {
            setError('Escolha uma senha forte que atenda a todos os requisitos')
            return
        }
        if (!passwordsMatch()) {
            setError('As senhas não coincidem')
            return
        }

        setLoading(true)
        try {
            await API.register(email(), password(), passwordConfirmation())
            props.onRegistered(email())
        } catch (error: unknown) {
            setError(errorMessage(error))
        } finally {
            setLoading(false)
        }
    }

    createEffect(() => {
        if (error()) {
            setTimeout(() => setError(''), 1000)
        }
    })

    return (
        <Center>
            <Card>
                <div>
                    <Title>Criar conta</Title>
                </div>

                <div style={{ display: 'flex', 'flex-direction': 'column', gap: '10px' }}>
                    <Input
                        label="Email"
                        type="email"
                        placeholder="voce@email.com"
                        value={email()}
                        onInput={e => setEmail(e.currentTarget.value)}
                    />
                    <Input
                        label="Senha"
                        type="password"
                        placeholder="Digite sua senha"
                        autocomplete="new-password"
                        value={password()}
                        onInput={e => setPassword(e.currentTarget.value)}
                        onKeyDown={e => e.key === 'Enter' && submit()}
                    />

                    <div
                        role="status"
                        aria-live="polite"
                        style={{ display: 'flex', 'flex-direction': 'column', gap: '6px', 'font-size': '11px' }}
                    >
                        <div style={{ display: 'flex', 'justify-content': 'space-between', color: 'var(--muted)' }}>
                            <span>Força da senha</span>
                            <span style={{ color: strength().tone === 'success' ? 'var(--success)' : strength().tone === 'error' ? 'var(--error)' : 'var(--muted)' }}>
                                {strength().label}
                            </span>
                        </div>
                        <div style={{ height: '5px', background: 'var(--border)', 'border-radius': '999px', overflow: 'hidden' }}>
                            <div
                                style={{
                                    height: '100%',
                                    width: `${(strength().score / strength().total) * 100}%`,
                                    background: strength().tone === 'success' ? 'var(--success)' : strength().tone === 'error' ? 'var(--error)' : 'var(--accent)',
                                    transition: 'width 0.15s, background 0.15s',
                                }}
                            />
                        </div>
                        <ul style={{ margin: '0', padding: '0', display: 'grid', 'grid-template-columns': 'repeat(2, minmax(0, 1fr))', gap: '3px 10px', 'list-style': 'none' }}>
                            <For each={strength().requirements}>{requirement => (
                                <li style={{ color: requirement.satisfied ? 'var(--success)' : 'var(--muted)' }}>
                                    {requirement.satisfied ? '✓' : '○'} {requirement.label}
                                </li>
                            )}</For>
                        </ul>
                    </div>

                    <Input
                        label="Confirmar senha"
                        type="password"
                        placeholder="Repita sua senha"
                        autocomplete="new-password"
                        value={passwordConfirmation()}
                        onInput={e => setPasswordConfirmation(e.currentTarget.value)}
                        onKeyDown={e => e.key === 'Enter' && submit()}
                    />
                    <Show when={passwordConfirmation()}>
                        <p style={{ margin: '-6px 0 0', 'font-size': '12px', color: passwordsMatch() ? 'var(--success)' : 'var(--error)' }}>
                            {passwordsMatch() ? '✓ As senhas coincidem' : 'As senhas não coincidem'}
                        </p>
                    </Show>
                </div>

                <Alert kind="error" message={error()} />

                <Button loading={loading()} onClick={submit}>
                    {loading() ? 'Criando conta…' : 'Criar conta'}
                </Button>

                <p style={{ 'font-size': '14px', color: 'var(--muted)', 'text-align': 'center' }}>
                    Já tem uma conta?{' '}
                    <TextLink onClick={props.onLogin}>Entrar</TextLink>
                </p>
            </Card>
        </Center>
    )
}
