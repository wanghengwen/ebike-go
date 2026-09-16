package com.luopingtech.ebike.rider.feature.riding

import com.luopingtech.ebike.rider.domain.riding.RidePhase
import com.luopingtech.ebike.rider.domain.riding.RideSession
import com.luopingtech.ebike.rider.domain.riding.RideStateMachine
import com.luopingtech.ebike.rider.platform.SecureStore
import kotlinx.serialization.json.Json

/**
 * 骑行相位的落盘 / 读回。
 *
 * 为什么必须持久化：C 端用户在骑行中会切到导航、接电话、拉后台，系统随时可能回收进程。
 * 只依赖 `getRideInfo` 恢复的话，冷启动那几百毫秒里界面无从判断该画首页还是骑行页 ——
 * 用户会看到首页闪一下再跳走，而更糟的情况是他在首页又扫了一辆车。
 *
 * 读回时一律过 [RideStateMachine.sanitizeRestored]：磁盘上的相位可能停在
 * `Unlocking` / `Returning` 这种「请求已发未确认」的窗口里。
 */
class RideSessionStore(
    private val secureStore: SecureStore,
    private val json: Json = DEFAULT_JSON,
) {
    fun load(): RideSession {
        val raw = secureStore.getString(SecureStore.KEY_RIDE_SESSION)
        if (raw.isNullOrBlank()) return RideSession.Idle
        val decoded = runCatching { json.decodeFromString(RideSession.serializer(), raw) }
            .getOrElse {
                // 词表升级导致的旧格式读不动时清掉，别让用户永远卡在坏快照上。
                clear()
                return RideSession.Idle
            }
        return RideStateMachine.sanitizeRestored(decoded)
    }

    fun save(session: RideSession) {
        if (session.phase == RidePhase.Idle && !session.hasVehicle) {
            clear()
            return
        }
        // Scanning 不落盘：相机会话本来就跨不过进程，存了只会在下次冷启动被清掉。
        val persistable = if (session.phase == RidePhase.Scanning) {
            session.copy(phase = RidePhase.Idle)
        } else {
            session
        }
        secureStore.putString(
            SecureStore.KEY_RIDE_SESSION,
            json.encodeToString(RideSession.serializer(), persistable),
        )
    }

    fun clear() {
        secureStore.remove(SecureStore.KEY_RIDE_SESSION)
    }

    var userPin: String
        get() = secureStore.getString(SecureStore.KEY_USER_PIN).orEmpty()
        set(value) {
            if (value.isBlank()) {
                secureStore.remove(SecureStore.KEY_USER_PIN)
            } else {
                secureStore.putString(SecureStore.KEY_USER_PIN, value)
            }
        }

    var serviceAreaId: String
        get() = secureStore.getString(SecureStore.KEY_SERVICE_AREA_ID).orEmpty()
        set(value) {
            if (value.isBlank()) {
                secureStore.remove(SecureStore.KEY_SERVICE_AREA_ID)
            } else {
                secureStore.putString(SecureStore.KEY_SERVICE_AREA_ID, value)
            }
        }

    /** 申诉成功后置位；回骑行页读一次就清（旧版 `autoLock` 语义）。 */
    fun markAutoReturn() {
        secureStore.putString(SecureStore.KEY_RIDE_AUTO_RETURN, "1")
    }

    fun consumeAutoReturn(): Boolean {
        val flagged = secureStore.getString(SecureStore.KEY_RIDE_AUTO_RETURN) == "1"
        if (flagged) secureStore.remove(SecureStore.KEY_RIDE_AUTO_RETURN)
        return flagged
    }

    companion object {
        private val DEFAULT_JSON = Json {
            ignoreUnknownKeys = true
            encodeDefaults = true
        }
    }
}
