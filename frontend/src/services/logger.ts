/**
 * vibe-coding-logger フロントエンド統合
 * ユーザーアクション、API呼び出し、エラーをログ出力
 */

type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR'

interface LogEntry {
  timestamp: string
  level: LogLevel
  message: string
  context?: Record<string, unknown>
}

class VibeCodingLogger {
  private enabled: boolean

  constructor() {
    this.enabled = import.meta.env.DEV ||
                   import.meta.env.VITE_VIBE_LOGGER_ENABLED === 'true'
  }

  private log(level: LogLevel, message: string, context?: Record<string, unknown>): void {
    if (!this.enabled) return

    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level,
      message,
      context,
    }

    // コンソールに出力
    const logMethod = level === 'ERROR' ? 'error' :
                      level === 'WARN' ? 'warn' :
                      level === 'DEBUG' ? 'debug' : 'info'
    console[logMethod](`[${entry.level}] ${entry.message}`, entry.context || '')
  }

  debug(message: string, context?: Record<string, unknown>): void {
    this.log('DEBUG', message, context)
  }

  info(message: string, context?: Record<string, unknown>): void {
    this.log('INFO', message, context)
  }

  warn(message: string, context?: Record<string, unknown>): void {
    this.log('WARN', message, context)
  }

  error(message: string, context?: Record<string, unknown>): void {
    this.log('ERROR', message, context)
  }

  /**
   * ユーザーアクション（クリック、入力等）をログ
   */
  logUserAction(action: string, context?: Record<string, unknown>): void {
    this.info(`ユーザーアクション: ${action}`, context)
  }

  /**
   * API呼び出しをログ
   */
  logApiCall(endpoint: string, method: string, context?: Record<string, unknown>): void {
    this.debug(`API呼び出し: ${method} ${endpoint}`, context)
  }

  /**
   * ページ遷移をログ
   */
  logPageTransition(from: string, to: string): void {
    this.info('ページ遷移', { from, to })
  }
}

export const logger = new VibeCodingLogger()
