import { logger } from '@/shared/logger'
import { bleSession } from './session'

export async function ensureBleAdapter(): Promise<boolean> {
  const ok = await bleSession.openAdapter()
  if (!ok) logger.warn('ble adapter unavailable')
  return ok
}
