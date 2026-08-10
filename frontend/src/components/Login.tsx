import { createEffect, createSignal } from 'solid-js'
import { useLocation } from '@solidjs/router'
import { Button, Card, Center, Title, TextLink } from '../styles'
import Input from '../fragments/Input'
import Alert from '../fragments/Alert'

import API from '../api/client'
import { errorMessage } from '../api/errorMessage'
import { safeReturnTo } from '../returnTo'


interface Props {
    onLoggedIn: (returnTo: string) => void
    onGoRegister: () => void
}

export default (props: Props) => {
    const location = useLocation()
    const [email, setEmail] = createSignal('')
    const [password, setPassword] = createSignal('')
    const [loading, setLoading] = createSignal(false)
    const [error, setError] = createSignal('')

    const submit = async () => {
        if (loading()) return

        setError('')
        setLoading(true)
        try {
            await API.login(email(), password())
            const params = new URLSearchParams(location.search)
            props.onLoggedIn(safeReturnTo(params.get('returnTo')))
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
                    <Title>Entrar</Title>
                </div>

                <div style={{ display: 'flex', 'flex-direction': 'column', gap: '12px' }}>
                    <Input
                        label="Email"
                        type="email"
                        placeholder="voce@email.com"
                        value={email()}
                        onInput={e => setEmail(e.target.value)}
                    />
                    <Input
                        label="Senha"
                        type="password"
                        placeholder="••••••••"
                        value={password()}
                        onInput={e => setPassword(e.currentTarget.value)}
                        onKeyDown={e => e.key === 'Enter' && submit()}
                    />
                </div>

                <Alert kind="error" message={error()} />

                <Button loading={loading()} onClick={submit}>
                    {loading() ? 'Entrando…' : 'Entrar'}
                </Button>

                <p style={{ 'font-size': '14px', color: 'var(--muted)', 'text-align': 'center' }}>
                    Não tem uma conta?{' '}
                    <TextLink onClick={props.onGoRegister}>Criar conta</TextLink>
                </p>
            </Card>
        </Center>
    )
}
