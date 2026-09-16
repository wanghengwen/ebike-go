package com.luopingtech.ebike.rider.data.auth

import com.luopingtech.ebike.rider.core.network.CommonRequestBody
import com.luopingtech.ebike.rider.core.network.SignedApiClient
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.platform.DeviceInfo

interface UserRemote {
    suspend fun fetchPersonInfo(): RiderResult<PersonInfoDto>
    suspend fun fetchUserAccount(): RiderResult<UserAccountDto>
}

class UserApi(
    private val signedApi: SignedApiClient,
    private val tenantIdProvider: () -> String,
    private val deviceInfo: DeviceInfo,
    private val deviceIdProvider: () -> String,
) : UserRemote {
    override suspend fun fetchPersonInfo(): RiderResult<PersonInfoDto> =
        signedApi.post(
            path = "client/user/user/personInfo",
            bodyJson = commonBody(),
            deserializer = PersonInfoDto.serializer(),
        )

    override suspend fun fetchUserAccount(): RiderResult<UserAccountDto> =
        signedApi.post(
            path = "client/ebike-account/user_account",
            bodyJson = commonBody(),
            deserializer = UserAccountDto.serializer(),
        )

    private fun commonBody(): String = CommonRequestBody.toJsonString(
        tenantId = tenantIdProvider(),
        deviceInfo = deviceInfo,
        deviceId = deviceIdProvider(),
    )
}
