const MAX_SOURCE_BYTES = 20 * 1024 * 1024
const TARGET_BYTES = 500 * 1024
const MAX_DIMENSION = 1200
const MIN_DIMENSION = 160
const QUALITIES = [0.86, 0.76, 0.66, 0.56, 0.46, 0.36, 0.28]
const ALLOWED_IMAGE_TYPES = new Set(['image/png', 'image/jpeg', 'image/webp', 'image/gif'])

function imageError(message: string) {
    return new Error(message)
}

function loadImage(file: File): Promise<HTMLImageElement> {
    return new Promise((resolve, reject) => {
        const objectUrl = URL.createObjectURL(file)
        const image = new Image()

        image.onload = () => {
            URL.revokeObjectURL(objectUrl)
            if (!image.naturalWidth || !image.naturalHeight) {
                reject(imageError('Não foi possível ler essa imagem.'))
                return
            }
            resolve(image)
        }
        image.onerror = () => {
            URL.revokeObjectURL(objectUrl)
            reject(imageError('Não foi possível ler essa imagem.'))
        }
        image.src = objectUrl
    })
}

function canvasBlob(canvas: HTMLCanvasElement, type: string, quality: number): Promise<Blob | null> {
    return new Promise(resolve => canvas.toBlob(resolve, type, quality))
}

function blobDataUrl(blob: Blob): Promise<string> {
    return new Promise((resolve, reject) => {
        const reader = new FileReader()
        reader.onload = () => {
            if (typeof reader.result === 'string') resolve(reader.result)
            else reject(imageError('Não foi possível preparar essa imagem.'))
        }
        reader.onerror = () => reject(imageError('Não foi possível preparar essa imagem.'))
        reader.readAsDataURL(blob)
    })
}

async function encode(canvas: HTMLCanvasElement, quality: number) {
    const webp = await canvasBlob(canvas, 'image/webp', quality)
    if (webp?.type === 'image/webp') return webp
    return canvasBlob(canvas, 'image/jpeg', quality)
}

export async function compressProfileImage(file: File): Promise<string> {
    if (!ALLOWED_IMAGE_TYPES.has(file.type)) {
        throw imageError('Escolha uma imagem PNG, JPEG, WEBP ou GIF.')
    }
    if (file.size > MAX_SOURCE_BYTES) {
        throw imageError('A imagem original deve ter no máximo 20 MB.')
    }

    const image = await loadImage(file)
    const canvas = document.createElement('canvas')
    const longestSide = Math.max(image.naturalWidth, image.naturalHeight)
    const initialScale = Math.min(1, MAX_DIMENSION / longestSide)
    let width = Math.max(1, Math.round(image.naturalWidth * initialScale))
    let height = Math.max(1, Math.round(image.naturalHeight * initialScale))

    try {
        for (let attempt = 0; attempt < 12; attempt += 1) {
            canvas.width = width
            canvas.height = height
            const context = canvas.getContext('2d')
            if (!context) throw imageError('Seu navegador não conseguiu preparar a imagem.')

            context.clearRect(0, 0, width, height)
            context.drawImage(image, 0, 0, width, height)

            for (const quality of QUALITIES) {
                const blob = await encode(canvas, quality)
                if (blob && blob.size <= TARGET_BYTES) return blobDataUrl(blob)
            }

            if (width <= MIN_DIMENSION && height <= MIN_DIMENSION) break
            width = Math.max(MIN_DIMENSION, Math.round(width * 0.82))
            height = Math.max(MIN_DIMENSION, Math.round(height * 0.82))
        }
    } finally {
        canvas.width = 1
        canvas.height = 1
    }

    throw imageError('Não foi possível comprimir a imagem para 512 KB.')
}
