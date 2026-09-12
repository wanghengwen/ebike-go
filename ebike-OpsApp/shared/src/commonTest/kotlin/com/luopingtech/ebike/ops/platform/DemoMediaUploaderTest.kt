package com.luopingtech.ebike.ops.platform

import com.luopingtech.ebike.ops.core.result.OpsResult
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking

class DemoMediaUploaderTest {
    @Test
    fun rewritesLocalAndDemoUrls() = runBlocking {
        val uploader = DemoMediaUploader()
        val result = uploader.upload(
            listOf(
                "demo://photo/1",
                "https://cdn.example/a.jpg",
                "content://media/1",
            ),
        )
        assertTrue(result is OpsResult.Ok)
        val urls = (result as OpsResult.Ok).value
        assertEquals("https://demo.cdn.ops/photo/1", urls[0])
        assertEquals("https://cdn.example/a.jpg", urls[1])
        assertTrue(urls[2].startsWith("https://demo.cdn.ops/upload/"))
    }
}
