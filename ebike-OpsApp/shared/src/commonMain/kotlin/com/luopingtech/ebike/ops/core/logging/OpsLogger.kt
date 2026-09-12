package com.luopingtech.ebike.ops.core.logging

enum class LogLevel { DEBUG, INFO, WARN, ERROR }

interface OpsLogger {
    fun log(level: LogLevel, tag: String, message: String, throwable: Throwable? = null)

    fun d(tag: String, message: String) = log(LogLevel.DEBUG, tag, message)
    fun i(tag: String, message: String) = log(LogLevel.INFO, tag, message)
    fun w(tag: String, message: String, throwable: Throwable? = null) =
        log(LogLevel.WARN, tag, message, throwable)

    fun e(tag: String, message: String, throwable: Throwable? = null) =
        log(LogLevel.ERROR, tag, message, throwable)
}

object NoOpLogger : OpsLogger {
    override fun log(level: LogLevel, tag: String, message: String, throwable: Throwable?) = Unit
}

/**
 * Simple stdout logger for local / demo builds. Host apps should replace with platform loggers.
 */
object StdoutLogger : OpsLogger {
    override fun log(level: LogLevel, tag: String, message: String, throwable: Throwable?) {
        val line = "[$level][$tag] $message"
        println(line)
        throwable?.printStackTrace()
    }
}
