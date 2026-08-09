import { execFileSync, spawn, spawnSync } from 'node:child_process'
import { accessSync, constants } from 'node:fs'
import { fileURLToPath } from 'node:url'

const frontendRoot = fileURLToPath(new URL('..', import.meta.url))
const viteBin = fileURLToPath(new URL('../node_modules/vite/bin/vite.js', import.meta.url))
const port = 5174
const url = `http://127.0.0.1:${port}/scripts/component-smoke.html`

const findBrowser = () => {
    const candidates = [
        process.env.CHROME_BIN,
        '/usr/bin/google-chrome',
        '/opt/google/chrome/google-chrome',
        'google-chrome',
        'chromium',
        'chromium-browser',
    ]
    for (const candidate of candidates) {
        if (!candidate) continue
        if (candidate.startsWith('/')) {
            try {
                accessSync(candidate, constants.X_OK)
                return candidate
            } catch {
                continue
            }
        }
        try {
            return execFileSync('which', [candidate], { encoding: 'utf8' }).trim()
        } catch {
            // Tenta o próximo nome conhecido.
        }
    }
    return null
}

const browser = findBrowser()
if (!browser) {
    throw new Error('Chrome/Chromium não encontrado; defina CHROME_BIN para executar os testes de componentes')
}

const vite = spawn(process.execPath, [viteBin, '--host', '127.0.0.1', '--port', String(port)], {
    cwd: frontendRoot,
    stdio: ['ignore', 'pipe', 'pipe'],
})
let viteOutput = ''
let viteError = null
vite.stdout.on('data', chunk => { viteOutput += chunk.toString() })
vite.stderr.on('data', chunk => { viteOutput += chunk.toString() })
vite.on('error', error => { viteError = error })

const waitForServer = async () => {
    for (let attempt = 0; attempt < 50; attempt += 1) {
        try {
            await fetch(url)
            return
        } catch {
            if (vite.exitCode !== null) {
                throw new Error(`Vite encerrou antes do teste: ${viteError ?? ''} ${viteOutput}`)
            }
            await new Promise(resolve => setTimeout(resolve, 100))
        }
    }
    throw new Error(`Vite não iniciou a tempo: ${viteError ?? ''} ${viteOutput}`)
}

try {
    await waitForServer()
    const result = spawnSync(browser, [
        '--headless=new',
        '--no-sandbox',
        '--disable-gpu',
        '--disable-dev-shm-usage',
        '--virtual-time-budget=5000',
        '--dump-dom',
        url,
    ], { encoding: 'utf8', maxBuffer: 4 * 1024 * 1024 })

    if (result.error) throw result.error
    if (result.status !== 0) {
        throw new Error(`Chrome encerrou com status ${result.status}: ${result.stderr}`)
    }
    if (!result.stdout.includes('data-test-status="passed"')) {
        throw new Error(`testes de componentes falharam: ${result.stdout}`)
    }
    process.stdout.write('component tests passed\n')
} finally {
    vite.kill('SIGTERM')
}
