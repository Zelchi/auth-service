export type PasswordRequirement = {
    label: string
    satisfied: boolean
}

export type PasswordStrength = {
    label: 'Digite sua senha' | 'Fraca' | 'Média' | 'Forte'
    score: number
    total: number
    isStrong: boolean
    tone: 'neutral' | 'error' | 'warning' | 'success'
    requirements: PasswordRequirement[]
}

const characterLength = (value: string) => Array.from(value).length

export const getPasswordStrength = (password: string): PasswordStrength => {
    const requirements: PasswordRequirement[] = [
        { label: '8 a 72 caracteres', satisfied: characterLength(password) >= 8 && characterLength(password) <= 72 },
        { label: 'uma letra minúscula', satisfied: /\p{Ll}/u.test(password) },
        { label: 'uma letra maiúscula', satisfied: /\p{Lu}/u.test(password) },
        { label: 'um número', satisfied: /\p{Nd}/u.test(password) },
    ]
    const score = requirements.filter(requirement => requirement.satisfied).length
    const isStrong = score === requirements.length

    if (!password) {
        return { label: 'Digite sua senha', score, total: requirements.length, isStrong, tone: 'neutral', requirements }
    }
    if (isStrong) {
        return { label: 'Forte', score, total: requirements.length, isStrong, tone: 'success', requirements }
    }
    if (score >= 3) {
        return { label: 'Média', score, total: requirements.length, isStrong, tone: 'warning', requirements }
    }
    return { label: 'Fraca', score, total: requirements.length, isStrong, tone: 'error', requirements }
}
