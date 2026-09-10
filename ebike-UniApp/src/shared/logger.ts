type Level = 'debug' | 'info' | 'warn' | 'error'

function emit(level: Level, ...args: unknown[]) {
  const prefix = `[luopingtech][${level}]`
  // eslint-disable-next-line no-console
  console[level === 'debug' ? 'log' : level](prefix, ...args)
}

export const logger = {
  debug: (...args: unknown[]) => emit('debug', ...args),
  info: (...args: unknown[]) => emit('info', ...args),
  warn: (...args: unknown[]) => emit('warn', ...args),
  error: (...args: unknown[]) => emit('error', ...args),
}
