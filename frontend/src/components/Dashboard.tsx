import { createResource, Show } from 'solid-js'
import { styled } from 'solid-styled-components'
import { Button } from '../styles'

import Cookies from 'js-cookie'
import API from '../api/client'

const Shell = styled('div')`
    min-height: 100dvh;
    width: 100dvw;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 24px;
    gap: 32px;
`

const ProfileCard = styled('div')`
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 16px;
    padding: 36px 32px;
    width: 100%;
    max-width: 400px;
    display: flex;
    flex-direction: column;
    gap: 24px;
`

const Avatar = styled('div')`
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: var(--accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 22px;
    font-weight: 700;
    color: var(--bg);

    @media (prefers-color-scheme: light) {
        color: #fff;
    }
`

const Field = styled('div')`
    display: flex;
    flex-direction: column;
    gap: 4px;
`

const FieldLabel = styled('span')`
    font-size: 12px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 0.08em;
`

const FieldValue = styled('span')`
    font-size: 15px;
    color: var(--text);
`

interface Props {
    onLogout: () => void
}

export default (props: Props) => {
    const [user] = createResource(() => API.me())

    const logout = () => {
        Cookies.remove('auth-token')
        props.onLogout()
    }

    const formatDate = (iso: string) =>
        new Date(iso).toLocaleDateString('pt-BR', {
            day: '2-digit', month: 'long', year: 'numeric',
        })

    return (
        <Shell>
            <Show when={user()} fallback={
                <p style={{ color: 'var(--muted)' }}>Carregando…</p>
            }>
                {data => (
                    <ProfileCard>
                        <div style={{ display: 'flex', 'align-items': 'center', gap: '16px' }}>
                            <Avatar>{data().email[0].toUpperCase()}</Avatar>
                            <div style={{ display: 'flex', 'flex-direction': 'column', gap: '4px' }}>
                                <span style={{ 'font-size': '16px', 'font-weight': '600', color: 'var(--text)' }}>
                                    Minha conta
                                </span>
                            </div>
                        </div>

                        <div style={{ display: 'flex', 'flex-direction': 'column', gap: '16px' }}>
                            <Field>
                                <FieldLabel>Email</FieldLabel>
                                <FieldValue>{data().email}</FieldValue>
                            </Field>
                            <Field>
                                <FieldLabel>ID</FieldLabel>
                                <FieldValue style={{ 'font-size': '13px', 'word-break': 'break-all', color: 'var(--muted)' }}>
                                    {data().id}
                                </FieldValue>
                            </Field>
                            <Field>
                                <FieldLabel>Membro desde</FieldLabel>
                                <FieldValue>{formatDate(data().created_at)}</FieldValue>
                            </Field>
                        </div>

                        <Button onClick={logout} style={{ background: 'transparent', border: '1px solid var(--border)', color: 'var(--muted)' }}>
                            Sair
                        </Button>
                    </ProfileCard>
                )}
            </Show>
        </Shell>
    )
}