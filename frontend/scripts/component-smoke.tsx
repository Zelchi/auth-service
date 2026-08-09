import { render } from 'solid-js/web'
import Login from '../src/components/Login'
import Register from '../src/components/Register'
import Verify from '../src/components/Verify'
import API from '../src/api/client'

const root = document.querySelector<HTMLDivElement>('#test-root')
if (!root) throw new Error('test root não encontrado')

const waitForUpdate = () => new Promise<void>(resolve => setTimeout(resolve, 0))

const assert = (condition: unknown, message: string) => {
    if (!condition) throw new Error(message)
}

const fill = async (element: HTMLInputElement, value: string) => {
    element.value = value
    element.dispatchEvent(new Event('input', { bubbles: true }))
    await waitForUpdate()
}

const resetRoot = () => {
    root.replaceChildren()
}

const testLogin = async () => {
    let loggedIn = false
    API.login = async () => ({ user_id: 'user-123' })

    const dispose = render(() => (
        <Login onLoggedIn={() => { loggedIn = true }} onGoRegister={() => undefined} />
    ), root)

    const inputs = root.querySelectorAll<HTMLInputElement>('input')
    await fill(inputs[0], 'user@example.com')
    await fill(inputs[1], 'correct-password')
    root.querySelectorAll<HTMLButtonElement>('button')[0].click()
    await waitForUpdate()

    assert(loggedIn, 'login não chamou onLoggedIn após resposta bem-sucedida')
    dispose()
    resetRoot()
}

const testRegister = async () => {
    let registeredEmail = ''
    API.register = async () => ({ message: 'código enviado' })

    const dispose = render(() => (
        <Register onRegistered={email => { registeredEmail = email }} onLogin={() => undefined} />
    ), root)

    const inputs = root.querySelectorAll<HTMLInputElement>('input')
    await fill(inputs[0], 'new@example.com')
    await fill(inputs[1], 'correct-password')
    root.querySelectorAll<HTMLButtonElement>('button')[0].click()
    await waitForUpdate()

    assert(registeredEmail === 'new@example.com', 'cadastro não repassou o email registrado')
    dispose()
    resetRoot()
}

const testVerify = async () => {
    let verifyCalled = false
    let resentEmail = ''
    API.verify = async () => {
        verifyCalled = true
        return { message: 'conta verificada' }
    }
    API.resend = async email => {
        resentEmail = email
        return { message: 'novo código enviado' }
    }

    const dispose = render(() => (
        <Verify email="pending@example.com" onVerified={() => { verified = true }} />
    ), root)

    const buttons = root.querySelectorAll<HTMLButtonElement>('button')
    buttons[1].click()
    await waitForUpdate()
    assert(resentEmail === 'pending@example.com', 'reenvio não usou o email pendente')

    const input = root.querySelector<HTMLInputElement>('input')
    await fill(input, '123456')
    buttons[0].click()
    await waitForUpdate()
    assert(verifyCalled, 'verificação não chamou a API após o envio do código')

    dispose()
    resetRoot()
}

try {
    await testLogin()
    await testRegister()
    await testVerify()
    document.body.dataset.testStatus = 'passed'
    document.body.textContent = 'component tests passed'
} catch (error) {
    document.body.dataset.testStatus = 'failed'
    document.body.dataset.testError = error instanceof Error ? error.message : String(error)
    document.body.textContent = document.body.dataset.testError
}
