import { createEffect, createSignal } from 'solid-js'
import { Button, Card, Center, Title, TextLink } from '../styles'
import Input from '../fragments/Input'
import Alert from '../fragments/Alert'

import API from '../api/client'

interface Props {
    onRegistered: (email: string) => void
    onLogin: () => void
}

export default (props: Props) => {
    const [email, setEmail] = createSignal('')
    const [password, setPassword] = createSignal('')
    const [loading, setLoading] = createSignal(false)
    const [error, setError] = createSignal('')

    const submit = async () => {
        setError('')
        setLoading(true)
        try {
            await API.register(email(), password())
            props.onRegistered(email())
        } catch (e: any) {
            setError(e.message)
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

                <div style={{ display: 'flex', 'flex-direction': 'column', gap: '12px' }}>
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
                        placeholder="Mínimo 8 caracteres"
                        value={password()}
                        onInput={e => setPassword(e.currentTarget.value)}
                        onKeyDown={e => e.key === 'Enter' && submit()}
                    />
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