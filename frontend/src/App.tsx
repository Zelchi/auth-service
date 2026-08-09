import { Router, Route, useNavigate } from "@solidjs/router";
import { lazy, onCleanup, onMount } from 'solid-js'
import API from './api/client'

import { installBridge } from './scripts/bridge'
import { clearPendingEmail, readPendingEmail, writePendingEmail } from './pendingStorage.mjs'

const Register = lazy(() => import('./components/Register'))
const Verify = lazy(() => import('./components/Verify'))
const Login = lazy(() => import('./components/Login'))
const Dashboard = lazy(() => import('./components/Dashboard'))

type Screen = 'login' | 'register' | 'verify' | 'dashboard'

const initialScreen = async (): Promise<Screen> => {
	try {
		await API.me()
		return 'dashboard'
	} catch {
		// Sem sessão válida, continua avaliando o cadastro pendente local.
	}

    return readPendingEmail(localStorage) ? 'verify' : 'login'
}

const setPendingEmail = (email: string) => {
	writePendingEmail(localStorage, email)
}

const InitialRedirect = () => {
    const navigate = useNavigate()

    onMount(() => {
        initialScreen()
            .then(screen => navigate(`/${screen}`, { replace: true }))
            .catch(() => navigate('/login', { replace: true }))
    })

    return <p style={{ color: 'var(--muted)', padding: '24px' }}>Carregando…</p>
}

const AppRoutes = () => {
    return (
        <>
            <Route
                path="/register"
                component={RegisterRoute}
            />
            <Route
                path="/verify"
                component={VerifyRoute}
            />
            <Route
                path="/login"
                component={LoginRoute}
            />
            <Route
                path="/dashboard"
                component={DashboardRoute}
            />
            <Route
                path="*"
                component={InitialRedirect}
            />
        </>
    )
}

const RegisterRoute = () => {
    const navigate = useNavigate()

    return <Register
        onRegistered={(email) => { setPendingEmail(email); navigate('/verify') }}
        onLogin={() => navigate('/login')}
    />
}

const VerifyRoute = () => {
    const navigate = useNavigate()

    return <Verify
        email={readPendingEmail(localStorage)}
        onVerified={() => { clearPendingEmail(localStorage); navigate('/login') }}
    />
}

const LoginRoute = () => {
    const navigate = useNavigate()

    return <Login
        onLoggedIn={() => navigate('/dashboard')}
        onGoRegister={() => navigate('/register')}
    />
}

const DashboardRoute = () => {
    const navigate = useNavigate()

    return <Dashboard onLogout={() => navigate('/login')} />
}

export default () => {
    onMount(() => {
        const uninstallBridge = installBridge()
        onCleanup(uninstallBridge)
    })

    return (
        <Router>
            <AppRoutes />
        </Router>
    )
}
