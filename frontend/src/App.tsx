import { Router, Route } from "@solidjs/router";
import { lazy, onCleanup } from 'solid-js'
import Cookies from 'js-cookie'
import API from './api/client'

import { installBridge } from './scripts/bridge'

const Register = lazy(() => import('./components/Register'))
const Verify = lazy(() => import('./components/Verify'))
const Login = lazy(() => import('./components/Login'))
const Dashboard = lazy(() => import('./components/Dashboard'))

type Screen = 'login' | 'register' | 'verify' | 'dashboard'

const verifyPeriod = (old: number): boolean => {
    const period = 15 * 60 * 1000;
    const now = Date.now();
    return now - old > period;
}

const initialScreen = async (): Promise<Screen> => {
    const token = Cookies.get('auth-token')
    if (token) {
        try {
            await API.me()
            return 'dashboard'
        } catch {
            Cookies.remove('auth-token')
        }
    }

    const pending = localStorage.getItem('pending-email')
    if (pending) {
        const [time] = pending.split('|')

        if (!verifyPeriod(Number(time))) {
            return 'verify'
        }

        localStorage.removeItem('pending-email')
    }

    return 'login'
}

const setPendingEmail = (email: string) => {
    localStorage.setItem('pending-email', `${Date.now()}|${email}`)
}

const getPendingEmail = () => {
    const pendingEmail = localStorage.getItem('pending-email')
    if (!pendingEmail) {
        return ''
    }

    const [time, email] = pendingEmail.split('|')
    if (verifyPeriod(Number(time))) {
        localStorage.removeItem('pending-email')
        return ''
    }

    return email || ''
}

const redirect = (screen: Screen) => {
    switch (screen) {
        case 'login': return window.location.href = '/login';
        case 'register': return window.location.href = '/register';
        case 'verify': return window.location.href = '/verify';
        case 'dashboard': return window.location.href = '/dashboard';
    }
}

export default () => {
    onCleanup(installBridge())
    return (
        <Router>
            <Route
                path="/register"
                component={() => <Register
                    onRegistered={(email) => { setPendingEmail(email); redirect('verify') }}
                    onLogin={() => redirect('login')}
                />}
            />
            <Route
                path="/verify"
                component={() => <Verify
                    email={getPendingEmail()}
                    onVerified={() => { localStorage.removeItem('pending-email'); redirect('login') }}
                />}
            />
            <Route
                path="/login"
                component={() => <Login
                    onLoggedIn={() => redirect('dashboard')}
                    onGoRegister={() => redirect('register')}
                />}
            />
            <Route
                path="/dashboard"
                component={() => <Dashboard
                    onLogout={() => redirect('login')}
                />}
            />
            <Route
                path="*"
                component={() => {
                    initialScreen().then(redirect)
                }}
            />
        </Router>
    )
}