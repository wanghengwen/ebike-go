package com.luopingtech.ebike.ops.core.i18n

import com.luopingtech.ebike.ops.platform.SecureStore
import kotlin.concurrent.Volatile
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

interface StringResolver {
    val language: OpsLanguage
    val acceptLanguage: String
    fun t(key: Str, vararg args: Any?): String
}

/**
 * Process-wide resolver used by shared features and network defaults.
 * Host apps replace via [OpsI18n.install] at startup.
 */
object Strings : StringResolver {
    @Volatile
    private var delegate: StringResolver = OpsI18n.fallback()

    fun install(resolver: StringResolver) {
        delegate = resolver
        LocaleContext.acceptLanguage = resolver.acceptLanguage
    }

    override val language: OpsLanguage get() = delegate.language
    override val acceptLanguage: String get() = delegate.acceptLanguage
    override fun t(key: Str, vararg args: Any?): String = delegate.t(key, *args)
}

/** Mutable Accept-Language for request builders that cannot take a lambda yet. */
object LocaleContext {
    @Volatile
    var acceptLanguage: String = OpsLanguage.ZH_CN.acceptLanguage
}

class OpsI18n(
    private val secureStore: SecureStore? = null,
    initial: OpsLanguage = OpsLanguage.ZH_CN,
) : StringResolver {
    private val _language = MutableStateFlow(initial)
    val languageFlow: StateFlow<OpsLanguage> = _language.asStateFlow()

    override val language: OpsLanguage get() = _language.value
    override val acceptLanguage: String get() = language.acceptLanguage

    init {
        LocaleContext.acceptLanguage = acceptLanguage
    }

    fun setLanguage(language: OpsLanguage) {
        _language.value = language
        secureStore?.putString(SecureStore.KEY_LANGUAGE, language.tag)
        LocaleContext.acceptLanguage = language.acceptLanguage
        Strings.install(this)
    }

    override fun t(key: Str, vararg args: Any?): String {
        val template = StringCatalogs.catalog(language)[key]
            ?: StringCatalogs.catalog(OpsLanguage.ZH_CN)[key]
            ?: key.name
        return format(template, args)
    }

    companion object {
        fun fallback(language: OpsLanguage = OpsLanguage.ZH_CN): OpsI18n = OpsI18n(initial = language)

        fun fromStore(
            secureStore: SecureStore,
            systemLanguage: String? = null,
        ): OpsI18n {
            val stored = secureStore.getString(SecureStore.KEY_LANGUAGE)
            // Explicit user choice wins; first launch follows the device language.
            val lang = when {
                !stored.isNullOrBlank() -> OpsLanguage.fromTag(stored)
                else -> OpsLanguage.fromSystemLanguage(systemLanguage ?: platformLanguageTag())
            }
            return OpsI18n(secureStore = secureStore, initial = lang).also { Strings.install(it) }
        }

        private fun format(template: String, args: Array<out Any?>): String {
            if (args.isEmpty()) return template
            var out = template
            args.forEachIndexed { index, arg ->
                out = out.replace("{$index}", arg?.toString().orEmpty())
            }
            return out
        }
    }
}
