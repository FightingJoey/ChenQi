/** Soft UI sounds via Web Audio API — no external assets/CDN. */

const SOUND_KEY = 'chenqi_sound_enabled'

let ctx: AudioContext | null = null

export function isSoundEnabled(): boolean {
  const v = localStorage.getItem(SOUND_KEY)
  if (v === null) return true // default on
  return v !== '0' && v !== 'false'
}

export function setSoundEnabled(on: boolean) {
  localStorage.setItem(SOUND_KEY, on ? '1' : '0')
}

function getCtx(): AudioContext | null {
  if (typeof window === 'undefined') return null
  const AC = window.AudioContext || (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
  if (!AC) return null
  if (!ctx) ctx = new AC()
  return ctx
}

/** Resume audio context after a user gesture if suspended. */
export async function unlockAudio() {
  const c = getCtx()
  if (c && c.state === 'suspended') {
    try {
      await c.resume()
    } catch {
      /* ignore */
    }
  }
}

function tone(
  c: AudioContext,
  freq: number,
  start: number,
  dur: number,
  gainPeak = 0.12,
  type: OscillatorType = 'sine',
) {
  const osc = c.createOscillator()
  const g = c.createGain()
  osc.type = type
  osc.frequency.value = freq
  g.gain.setValueAtTime(0.0001, start)
  g.gain.exponentialRampToValueAtTime(gainPeak, start + 0.02)
  g.gain.exponentialRampToValueAtTime(0.0001, start + dur)
  osc.connect(g)
  g.connect(c.destination)
  osc.start(start)
  osc.stop(start + dur + 0.02)
}

/** Pleasant short chime when a SAVERS step completes. */
export function playStepComplete() {
  if (!isSoundEnabled()) return
  const c = getCtx()
  if (!c) return
  void unlockAudio()
  const t = c.currentTime
  tone(c, 523.25, t, 0.18, 0.1) // C5
  tone(c, 659.25, t + 0.12, 0.22, 0.11) // E5
  tone(c, 783.99, t + 0.24, 0.28, 0.08) // G5
}

/** Softer blip when skipping a step. */
export function playStepSkip() {
  if (!isSoundEnabled()) return
  const c = getCtx()
  if (!c) return
  void unlockAudio()
  const t = c.currentTime
  tone(c, 392.0, t, 0.12, 0.06, 'triangle') // G4
}

/** Slightly richer celebration when the full daily session is done. */
export function playSessionComplete() {
  if (!isSoundEnabled()) return
  const c = getCtx()
  if (!c) return
  void unlockAudio()
  const t = c.currentTime
  const notes = [523.25, 659.25, 783.99, 1046.5] // C5 E5 G5 C6
  notes.forEach((f, i) => {
    tone(c, f, t + i * 0.11, 0.32, 0.12 - i * 0.01)
  })
  // soft chord linger
  tone(c, 523.25, t + 0.45, 0.45, 0.05)
  tone(c, 659.25, t + 0.45, 0.45, 0.05)
  tone(c, 783.99, t + 0.45, 0.5, 0.04)
}
