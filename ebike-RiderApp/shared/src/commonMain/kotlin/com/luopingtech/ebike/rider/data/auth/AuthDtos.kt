package com.luopingtech.ebike.rider.data.auth

import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.KSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.descriptors.PrimitiveKind
import kotlinx.serialization.descriptors.PrimitiveSerialDescriptor
import kotlinx.serialization.descriptors.SerialDescriptor
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder
import kotlinx.serialization.json.JsonDecoder
import kotlinx.serialization.json.JsonNames
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.contentOrNull

@OptIn(ExperimentalSerializationApi::class)
@Serializable
data class OauthTokenDto(
    @SerialName("accessToken")
    @JsonNames("access_token")
    val accessToken: String = "",
    @SerialName("refreshToken")
    @JsonNames("refresh_token")
    val refreshToken: String = "",
    @SerialName("tokenType")
    @JsonNames("token_type")
    val tokenType: String = "Bearer",
    @SerialName("expiresIn")
    @JsonNames("expires_in")
    val expiresIn: Long = 0,
    val phone: String = "",
    val pin: String = "",
    val nickname: String = "",
    val avatar: String = "",
    val grantType: String = "",
    val jti: String = "",
)

@OptIn(ExperimentalSerializationApi::class)
@Serializable
data class PersonInfoDto(
    @Serializable(with = FlexibleStringSerializer::class)
    val pin: String = "",
    val phone: String = "",
    @SerialName("nickname")
    @JsonNames("nickName")
    val nickname: String = "",
    val avatar: String = "",
    val balance: Long = 0,
    val recharge: Long = 0,
    val present: Long = 0,
    @Serializable(with = FlexibleStringSerializer::class)
    val serviceId: String = "",
    val izAuth: Boolean = false,
    val izNeedAuth: Boolean = false,
    val ridingState: Int = 0,
    val payState: Int = 0,
    val izRiding: Boolean = false,
    val depositState: Int = 0,
    val needFaceCheck: Int = 0,
    val inBlacklist: Boolean = false,
    val authName: String = "",
    val description: String = "",
)

@Serializable
data class UserAccountDto(
    val userWallet: WalletDto? = null,
    @Serializable(with = FlexibleStringSerializer::class)
    val pin: String = "",
    val balance: Long = 0,
    val recharge: Long = 0,
    val present: Long = 0,
)

@Serializable
data class WalletDto(
    val balance: Long = 0,
    val recharge: Long = 0,
    val present: Long = 0,
    val depositedMount: Long = 0,
)

/** Backend often sends pin / serviceId as number or string. */
object FlexibleStringSerializer : KSerializer<String> {
    override val descriptor: SerialDescriptor =
        PrimitiveSerialDescriptor("FlexibleString", PrimitiveKind.STRING)

    override fun serialize(encoder: Encoder, value: String) {
        encoder.encodeString(value)
    }

    override fun deserialize(decoder: Decoder): String {
        val json = decoder as? JsonDecoder ?: return decoder.decodeString()
        return when (val element = json.decodeJsonElement()) {
            is JsonNull -> ""
            is JsonPrimitive -> element.contentOrNull.orEmpty()
            else -> element.toString()
        }
    }
}
