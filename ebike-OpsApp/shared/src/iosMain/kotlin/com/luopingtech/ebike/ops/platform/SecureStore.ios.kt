package com.luopingtech.ebike.ops.platform

actual fun createSecureStore(): SecureStore = InMemorySecureStore()
