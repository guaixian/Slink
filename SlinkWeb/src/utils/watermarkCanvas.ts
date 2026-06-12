export type WmPosition = 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'center'

export interface GroupPolicyLike {
  image_save_quality?: number
  is_enable_watermark?: number
  watermark_configs?: {
    driver?: string
    drivers?: Record<string, Record<string, unknown>>
  }
}

function parsePosition(v: unknown): WmPosition {
  const s = String(v ?? '').toLowerCase()
  if (['top-left', 'top-right', 'bottom-left', 'bottom-right', 'center'].includes(s)) {
    return s as WmPosition
  }
  return 'bottom-right'
}

function resolvePoint(
  w: number,
  h: number,
  ww: number,
  wh: number,
  pos: WmPosition,
  mx: number,
  my: number
) {
  switch (pos) {
    case 'top-left':
      return { x: mx, y: my }
    case 'top-right':
      return { x: w - ww - mx, y: my }
    case 'bottom-left':
      return { x: mx, y: h - wh - my }
    case 'center':
      return { x: (w - ww) / 2, y: (h - wh) / 2 }
    default:
      return { x: w - ww - mx, y: h - wh - my }
  }
}

function staticUrl(rel: string): string {
  const path = rel.replace(/^\/+/, '')
  if (/^https?:\/\//i.test(rel)) return rel
  return `${window.location.origin}/static/${path}`
}

async function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error(`无法加载图片: ${src}`))
    img.src = src
  })
}

async function ensureFontLoaded(fontFile: string, family: string): Promise<void> {
  const f = fontFile.trim()
  if (!f) return
  try {
    const url = staticUrl(f)
    const face = new FontFace(family, `url(${url})`)
    await face.load()
    document.fonts.add(face)
  } catch {
    /* 使用系统后备字体 */
  }
}

/**
 * 按全局策略中的 watermark_configs 在浏览器端叠加水印，返回可上传的 File。
 * GIF 仅处理首帧；无法加载水印图/字体时跳过对应图层不中断。
 */
export async function watermarkedFileFromPolicy(
  file: File,
  policy: GroupPolicyLike,
  textOverride?: string
): Promise<File> {
  const wc = policy.watermark_configs
  const driver = String(wc?.driver ?? 'font').toLowerCase()
  let q = Number(policy.image_save_quality) || 85
  if (q < 1) q = 85
  if (q > 100) q = 100

  const bitmap = await createImageBitmap(file)
  const w = bitmap.width
  const h = bitmap.height
  const canvas = document.createElement('canvas')
  canvas.width = w
  canvas.height = h
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('Canvas 不可用')
  ctx.drawImage(bitmap, 0, 0)
  bitmap.close()

  if (driver === 'image') {
    const d = wc?.drivers?.image
    if (d) {
      const rel = d.image != null ? String(d.image) : ''
      if (rel) {
        try {
          const wmImg = await loadImage(staticUrl(rel))
          let iw = wmImg.naturalWidth
          let ih = wmImg.naturalHeight
          const tw = Number(d.width) || 0
          const th = Number(d.height) || 0
          if (tw > 0 && th > 0) {
            iw = tw
            ih = th
          }
          let op = Number(d.opacity) || 100
          if (op <= 0) op = 100
          const pos = parsePosition(d.position)
          const mx = Number(d.x) || 0
          const my = Number(d.y) || 0
          const pt = resolvePoint(w, h, iw, ih, pos, mx, my)
          ctx.save()
          ctx.globalAlpha = Math.min(1, op / 100)
          ctx.drawImage(wmImg, pt.x, pt.y, iw, ih)
          ctx.restore()
        } catch {
          /* 水印图失败则跳过 */
        }
      }
    }
  } else {
    const d = wc?.drivers?.font
    if (d) {
      let text = textOverride != null && textOverride !== '' ? textOverride : String(d.text ?? '')
      if (text) {
        const color = String(d.color ?? '#ffffff')
        const size = Number(d.size) || 24
        const pos = parsePosition(d.position)
        const mx = Number(d.x) || 0
        const my = Number(d.y) || 0
        const fontFile = d.font != null ? String(d.font) : ''
        const family = 'CustomWmFont'
        await ensureFontLoaded(fontFile, family)

        ctx.save()
        ctx.textBaseline = 'top'
        ctx.font = `${size}px ${family}, sans-serif`
        ctx.fillStyle = color
        const metrics = ctx.measureText(text)
        const tw = metrics.width
        const th = size * 1.25
        const pt = resolvePoint(w, h, tw, th, pos, mx, my)
        ctx.fillText(text, pt.x, pt.y)
        ctx.restore()
      }
    }
  }

  const useJpeg = file.type === 'image/jpeg' || /\.jpe?g$/i.test(file.name)
  const mime = useJpeg ? 'image/jpeg' : 'image/png'
  const blob = await new Promise<Blob>((resolve, reject) => {
    canvas.toBlob(
      (b) => (b ? resolve(b) : reject(new Error('导出图片失败'))),
      mime,
      useJpeg ? q / 100 : undefined
    )
  })

  const base = file.name.replace(/\.[^.]+$/, '')
  const ext = useJpeg ? '.jpg' : '.png'
  return new File([blob], `${base}_wm${ext}`, { type: mime })
}
