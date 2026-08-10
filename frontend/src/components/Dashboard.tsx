import { createEffect, createResource, createSignal, Show } from 'solid-js'
import { styled } from 'solid-styled-components'
import { Button, Title } from '../styles'
import Input from '../fragments/Input'
import API from '../api/client'
import { compressProfileImage } from '../profileImage'
import { completeReturnTo, safeReturnTo } from '../returnTo'

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
    max-width: 440px;
    display: flex;
    flex-direction: column;
    gap: 20px;
`

const Avatar = styled('div')`
    width: 64px;
    height: 64px;
    flex: 0 0 64px;
    border-radius: 50%;
    background: #252525;
    border: 1px solid #3a3a3a;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    font-size: 24px;
    font-weight: 700;
    color: var(--text);

    img {
        width: 100%;
        height: 100%;
        object-fit: cover;
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

const FileInput = styled('input')`
    width: 100%;
    padding: 10px;
    color: var(--muted);
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    font-family: inherit;
    font-size: 13px;
`

const Hint = styled('p')`
    margin: -8px 0 0;
    color: var(--muted);
    font-size: 12px;
    line-height: 1.5;
`

const Message = styled('p') <{ kind: 'error' | 'success' }>`
    margin: -8px 0 0;
    color: ${p => p.kind === 'error' ? 'var(--error)' : 'var(--success)'};
    font-size: 13px;
`

interface Props {
    onLogout: () => void
    returnTo?: string
}

function initialFor(name: string, email: string) {
    const value = name.trim() || email.trim()
    return value ? value[0].toUpperCase() : '?'
}

export default (props: Props) => {
    const [user, { mutate }] = createResource(() => API.me())
    const [name, setName] = createSignal('')
    const [image, setImage] = createSignal('')
    const [saving, setSaving] = createSignal(false)
    const [compressing, setCompressing] = createSignal(false)
    const [error, setError] = createSignal('')
    const [success, setSuccess] = createSignal('')
    const [initialized, setInitialized] = createSignal(false)
    let redirected = false
    let returning = false
    let fileInput: HTMLInputElement | undefined

    createEffect(() => {
        const data = user()
        if (!data || initialized()) return

        setName(data.name)
        setImage(data.image)
        setInitialized(true)
    })

    createEffect(() => {
        if (user.error && !redirected) {
            redirected = true
            props.onLogout()
        }
    })

    createEffect(() => {
        const data = user()
        const target = props.returnTo ? safeReturnTo(props.returnTo) : ''
        if (!data?.name || !target || target === '/dashboard' || returning) return

        returning = true
        completeReturnTo(target)
    })

    const logout = async () => {
        try {
            await API.logout()
        } finally {
            props.onLogout()
        }
    }

    const selectImage = async (event: Event) => {
        const input = event.currentTarget as HTMLInputElement
        const file = input.files?.[0]
        if (!file) return

        setError('')
        setSuccess('')

        setCompressing(true)
        try {
            setImage(await compressProfileImage(file))
        } catch (compressionError: unknown) {
            setError(compressionError instanceof Error ? compressionError.message : 'Não foi possível comprimir a foto.')
            input.value = ''
        } finally {
            setCompressing(false)
        }
    }

    const removeImage = () => {
        setImage('')
        if (fileInput) fileInput.value = ''
    }

    const saveProfile = async () => {
        if (saving() || compressing()) return

        setError('')
        setSuccess('')
        setSaving(true)
        try {
            const updated = await API.updateProfile(name(), image())
            mutate(updated)
            setName(updated.name)
            setImage(updated.image)
            setSuccess('Perfil atualizado com sucesso.')
        } catch (saveError: unknown) {
            setError(saveError instanceof Error ? saveError.message : 'Não foi possível atualizar o perfil.')
        } finally {
            setSaving(false)
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
                            <Avatar>
                                <Show when={image()} fallback={initialFor(name(), data().email)}>
                                    <img alt="" src={image()} />
                                </Show>
                            </Avatar>
                            <div style={{ display: 'flex', 'flex-direction': 'column', gap: '4px' }}>
                                <span style={{ 'font-size': '18px', 'font-weight': '600', color: 'var(--text)' }}>
                                    {name() || 'Complete seu perfil'}
                                </span>
                                <span style={{ 'font-size': '13px', color: 'var(--muted)' }}>{data().email}</span>
                            </div>
                        </div>

                        <div style={{ display: 'flex', 'flex-direction': 'column', gap: '14px' }}>
                            <Show when={!data().name}>
                                <div>
                                    <Title style={{ 'font-size': '20px' }}>Como devemos chamar você?</Title>
                                    <Hint>Escolha um nome único. Você poderá alterá-lo depois.</Hint>
                                </div>
                            </Show>

                            <Input
                                label="Nome público"
                                maxlength={80}
                                placeholder="Seu nome"
                                value={name()}
                                onInput={event => setName(event.currentTarget.value)}
                            />

                            <div style={{ display: 'flex', 'flex-direction': 'column', gap: '6px' }}>
                                <label style={{ 'font-size': '13px', color: 'var(--muted)', 'letter-spacing': '0.03em' }}>
                                    Foto de perfil (opcional)
                                </label>
                                <FileInput
                                    accept="image/png,image/jpeg,image/webp,image/gif"
                                    disabled={saving() || compressing()}
                                    onChange={selectImage}
                                    ref={element => { fileInput = element }}
                                    type="file"
                                />
                                <Hint>Imagens grandes são redimensionadas e comprimidas automaticamente para até 512 KB. Sem foto, usamos a inicial do seu nome.</Hint>
                            </div>

                            <Show when={image()}>
                                <div style={{ display: 'flex', 'align-items': 'center', gap: '12px' }}>
                                    <Avatar>
                                        <img alt="Prévia da foto de perfil" src={image()} />
                                    </Avatar>
                                    <Button
                                        onClick={removeImage}
                                        style={{ background: 'transparent', border: '1px solid var(--border)', color: 'var(--muted)', 'margin-top': '0' }}
                                        type="button"
                                    >
                                        Remover foto
                                    </Button>
                                </div>
                            </Show>

                            <Show when={error()}>
                                <Message aria-live="polite" kind="error">{error()}</Message>
                            </Show>
                            <Show when={success()}>
                                <Message aria-live="polite" kind="success">{success()}</Message>
                            </Show>

                            <Button loading={saving() || compressing()} onClick={saveProfile} type="button">
                                {compressing() ? 'Processando foto…' : saving() ? 'Salvando…' : data().name ? 'Salvar alterações' : 'Continuar'}
                            </Button>
                        </div>

                        <Show when={data().name}>
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
                        </Show>

                        <Button onClick={logout} style={{ background: 'transparent', border: '1px solid var(--border)', color: 'var(--muted)' }} type="button">
                            Sair
                        </Button>
                    </ProfileCard>
                )}
            </Show>
        </Shell>
    )
}
