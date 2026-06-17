import { createSignal } from 'solid-js'
import { Button, Card, Center, Title, Subtitle } from '../styles'
import Input from '../fragments/Input'
import Alert from '../fragments/Alert'

import API from '../api/client'

interface Props {
    email: string
    onVerified: () => void
}

export default (props: Props) => {
    const [code, setCode] = createSignal('')
    const [loading, setLoading] = createSignal(false)
    const [error, setError] = createSignal('')
    const [success, setSuccess] = createSignal('')

    const submit = async () => {
        setError('')
        setSuccess('')
        setLoading(true)
        try {
            await API.verify(props.email, code())
            setSuccess('Conta verificada! Redirecionando…')
            setTimeout(() => props.onVerified(), 1200)
        } catch (e: any) {
            setError(e.message)
        } finally {
            setLoading(false)
        }
    }

    return (
        <Center>
            <Card>
                <div>
                    <Title>Verifique seu email</Title>
                    <Subtitle>
                        Enviamos um código de 6 dígitos para{' '}
                        <span style={{ color: 'var(--accent)' }}>{props.email}</span>.
                    </Subtitle>
                </div>

                <Input
                    label="Código de verificação"
                    type="text"
                    placeholder="000000"
                    maxLength={6}
                    value={code()}
                    onInput={e => setCode(e.currentTarget.value.replace(/\D/g, ''))}
                    onKeyDown={e => e.key === 'Enter' && submit()}
                    style={{
                        'font-size': '28px',
                        'letter-spacing': '10px',
                        'text-align': 'center',
                        'padding': '14px',
                    }}
                />

                <Alert kind="error" message={error()} />
                <Alert kind="success" message={success()} />

                <Button loading={loading()} onClick={submit}>
                    {loading() ? 'Verificando…' : 'Confirmar código'}
                </Button>

                <p style={{ 'font-size': '13px', color: 'var(--muted)', 'text-align': 'center' }}>
                    O código expira em 15 minutos.
                </p>
            </Card>
        </Center>
    )
}