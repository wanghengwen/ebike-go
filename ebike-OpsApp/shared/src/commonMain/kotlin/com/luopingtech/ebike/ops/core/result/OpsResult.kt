package com.luopingtech.ebike.ops.core.result

/**
 * Cross-platform operation result. UI layers map this to presentation state.
 */
sealed class OpsResult<out T> {
    data class Ok<T>(val value: T) : OpsResult<T>()
    data class Err(val error: OpsError) : OpsResult<Nothing>()

    val isOk: Boolean get() = this is Ok
    val isErr: Boolean get() = this is Err

    fun getOrNull(): T? = (this as? Ok)?.value

    inline fun <R> map(transform: (T) -> R): OpsResult<R> = when (this) {
        is Ok -> Ok(transform(value))
        is Err -> this
    }

    inline fun onSuccess(block: (T) -> Unit): OpsResult<T> {
        if (this is Ok) block(value)
        return this
    }

    inline fun onFailure(block: (OpsError) -> Unit): OpsResult<T> {
        if (this is Err) block(error)
        return this
    }
}

data class OpsError(
    val code: String,
    val message: String,
    val cause: Throwable? = null,
) {
    companion object {
        fun network(message: String, cause: Throwable? = null) =
            OpsError(code = "NETWORK", message = message, cause = cause)

        fun unauthorized(message: String = "unauthorized") =
            OpsError(code = "UNAUTHORIZED", message = message)

        fun business(code: String, message: String) =
            OpsError(code = code, message = message)

        fun cancelled(message: String = "cancelled") =
            OpsError(code = "CANCELLED", message = message)

        fun unsupported(capability: String) =
            OpsError(code = "UNSUPPORTED", message = "$capability is not available")
    }
}
