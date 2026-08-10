import { createEffect, createResource, Show } from 'solid-js'
import { styled } from 'solid-styled-components'
import { Button } from '../styles'

import API from '../api/client'

const Shell = styled('div')`
    min-height: 100dvh;
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 20px;
    gap: 24px;
`

const ProfileCard = styled('div')`
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 16px;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.32);
    padding: 24px;
    width: 100%;
    max-width: 380px;
    display: flex;
    flex-direction: column;
    gap: 20px;
`

const Avatar = styled('div')`
    width: 52px;
    height: 52px;
    border-radius: 50%;
    background: var(--accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 22px;
    font-weight: 700;
    color: var(--bg);
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
    let redirected = false

    createEffect(() => {
        if (user.error && !redirected) {
            redirected = true
            props.onLogout()
        }
    })

    const logout = async () => {
        try {
            await API.logout()
        } finally {
            props.onLogout()
        }
    }

    const formatDate = (iso: string) =>
        new Date(iso).toLocaleDateString('pt-BR', {
            day: '2-digit', month: 'long', year: 'numeric',
        })

    return (
        <Shell>
            <Show when={user()} fallback={
                <Show when={!user.error} fallback={
                    <p style={{ color: 'var(--muted)' }}>Sessão expirada. Redirecionando…</p>
                }>
                    <p style={{ color: 'var(--muted)' }}>Carregando…</p>
                </Show>
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