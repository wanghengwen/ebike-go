package com.luopingtech.ebike.rider.core.result

/**
 * Cross-platform operation result. UI layers map this to presentation state.
 */
sealed class RiderResult<out T> {
    data class Ok<T>(val value: T) : RiderResult<T>()
    data class Err(val error: RiderError) : RiderResult<Nothing>()

    val isOk: Boolean get() = this is Ok
    val isErr: Boolean get() = this is Err

    fun getOrNull(): T? = (this as? Ok)?.value

    inline fun <R> map(transform: (T) -> R): RiderResult<R> = when (this) {
        is Ok -> Ok(transform(value))
        is Err -> this
    }

    inline fun onSuccess(block: (T) -> Unit): RiderResult<T> {
        if (this is Ok) block(value)
        return this
    }

    inline fun onFailure(block: (RiderError) -> Unit): RiderResult<T> {
        if (this is Err) block(error)
        return this
    }
}

data class RiderError(
    val code: String,
    val message: String,
    val cause: Throwable? = null,
) {
    companion object {
        fun network(message: String, cause: Throwable? = null) =
            RiderError(code = "NETWORK", message = message, cause = cause)

        fun unauthorized(message: String = "unauthorized") =
            RiderError(code = "UNAUTHORIZED", message = message)

        fun business(code: String, message: String) =
            RiderError(code = code, message = message)

        fun cancelled(message: String = "cancelled") =
            RiderError(code = "CANCELLED", message = message)

        fun unsupported(capability: String) =
            RiderError(code = "UNSUPPORTED", message = "$capability is not available")
    }
}
