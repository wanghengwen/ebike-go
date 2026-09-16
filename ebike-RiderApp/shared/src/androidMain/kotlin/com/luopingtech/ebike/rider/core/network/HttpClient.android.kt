package com.luopingtech.ebike.rider.core.network

import io.ktor.client.HttpClient
import io.ktor.client.engine.okhttp.OkHttp

actual fun createPlatformHttpClient(): HttpClient = HttpClient(OkHttp)
