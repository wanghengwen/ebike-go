import { rentCheck } from '@/api/user'

/** Scan gate: face / rent eligibility before unlock */
export { rentCheck }

export function scanRentCheck(data: Record<string, unknown>) {
  return rentCheck(data)
}
