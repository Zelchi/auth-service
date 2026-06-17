import { Router, Route } from "@solidjs/router";
import { lazy } from 'solid-js'
import Cookies from 'js-cookie'
import API from './api/client'

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

export default () => {
    const initialScreen = (): Screen => {
        if (Cookies.get('auth-token')) API.me().then(() => redirect('dashboard'))

        if (localStorage.getItem('pending-email')) {
            const [time] = localStorage.getItem('pending-email')!.split('|')
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
        if (pendingEmail) {
            localStorage.removeItem('pending-email')
        }
        return pendingEmail ? pendingEmail.split('|')[1] : ''
    }

    const redirect = (screen: Screen) => {
        switch (screen) {
            case 'login': return window.location.href = '/login';
            case 'register': return window.location.href = '/register';
            case 'verify': return window.location.href = '/verify';
            case 'dashboard': return window.location.href = '/dashboard';
        }
    }

    return (
        <Router>
            <Route path="/register" component={() => <Register onRegistered={(email) => { setPendingEmail(email); redirect('verify') }} onLogin={() => redirect('login')} />} />
            <Route path="/verify" component={() => <Verify email={getPendingEmail()} onVerified={() => redirect('login')} />} />
            <Route path="/login" component={() => <Login onLoggedIn={() => redirect('dashboard')} onGoRegister={() => redirect('register')} />} />
            <Route path="/dashboard" component={() => <Dashboard onLogout={() => redirect('login')} />} />
            <Route path="*" component={() => redirect(initialScreen())} />
        </Router>
    )
}