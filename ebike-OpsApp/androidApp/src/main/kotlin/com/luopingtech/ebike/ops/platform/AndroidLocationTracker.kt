package com.luopingtech.ebike.ops.platform

import android.Manifest
import android.annotation.SuppressLint
import android.content.Context
import android.content.pm.PackageManager
import android.location.Location
import android.location.LocationListener
import android.location.LocationManager
import android.os.Bundle
import android.os.Looper
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlin.coroutines.resume

/**
 * System [LocationManager] tracker (GPS preferred, network fallback).
 * Matches legacy domestic cadence: minTime≈5s, minDistance=0.
 * Does not depend on Google Play Services.
 *
 * 保活 FGS 对齐原版 APPKeepService：由 [MainActivity] 在退后台时拉起、回前台时停掉，
 * 这里只标记是否在采点，不直接 startForeground。
 */
class AndroidLocationTracker(
    context: Context,
) : LocationTracker {
    private val appContext = context.applicationContext
    private val locationManager =
        appContext.getSystemService(Context.LOCATION_SERVICE) as LocationManager

    /** 当前是否在采点上报（登录后 TrackUploadFeature 打开）。 */
    @Volatile
    var trackingActive: Boolean = false
        private set

    override fun startTracking() {
        if (!hasPermission()) return
        trackingActive = true
    }

    override fun stopTracking() {
        trackingActive = false
        TrackLocationService.stop(appContext)
    }

    override suspend fun currentLocation(): OpsResult<GeoPoint> {
        if (!hasPermission()) {
            return OpsResult.Err(OpsError.business("LOC_PERM", "location permission required"))
        }
        val last = lastKnown()
        if (last != null) return OpsResult.Ok(last.toGeo())
        return requestSingleUpdate()
    }

    override fun lastKnownOrNull(): GeoPoint? = lastKnown()?.toGeo()

    override fun track(): Flow<GeoPoint> = callbackFlow {
        if (!hasPermission()) {
            close(IllegalStateException("location permission required"))
            return@callbackFlow
        }
        val listener = object : LocationListener {
            override fun onLocationChanged(location: Location) {
                trySend(location.toGeo())
            }

            @Deprecated("Deprecated in Java")
            override fun onStatusChanged(provider: String?, status: Int, extras: Bundle?) = Unit

            override fun onProviderEnabled(provider: String) = Unit
            override fun onProviderDisabled(provider: String) = Unit
        }
        @SuppressLint("MissingPermission")
        fun register(provider: String) {
            if (!locationManager.isProviderEnabled(provider)) return
            locationManager.requestLocationUpdates(
                provider,
                5_000L,
                0f,
                listener,
                Looper.getMainLooper(),
            )
        }
        register(LocationManager.GPS_PROVIDER)
        register(LocationManager.NETWORK_PROVIDER)
        lastKnown()?.let { trySend(it.toGeo()) }
        awaitClose {
            runCatching { locationManager.removeUpdates(listener) }
        }
    }

    fun hasPermission(): Boolean {
        val fine = appContext.checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION) ==
            PackageManager.PERMISSION_GRANTED
        val coarse = appContext.checkSelfPermission(Manifest.permission.ACCESS_COARSE_LOCATION) ==
            PackageManager.PERMISSION_GRANTED
        return fine || coarse
    }

    @SuppressLint("MissingPermission")
    private fun lastKnown(): Location? {
        if (!hasPermission()) return null
        val gps = runCatching {
            locationManager.getLastKnownLocation(LocationManager.GPS_PROVIDER)
        }.getOrNull()
        val net = runCatching {
            locationManager.getLastKnownLocation(LocationManager.NETWORK_PROVIDER)
        }.getOrNull()
        return listOfNotNull(gps, net).maxByOrNull { it.time }
    }

    @SuppressLint("MissingPermission")
    private suspend fun requestSingleUpdate(): OpsResult<GeoPoint> =
        suspendCancellableCoroutine { cont ->
            if (!hasPermission()) {
                cont.resume(OpsResult.Err(OpsError.business("LOC_PERM", "location permission required")))
                return@suspendCancellableCoroutine
            }
            val provider = when {
                locationManager.isProviderEnabled(LocationManager.GPS_PROVIDER) ->
                    LocationManager.GPS_PROVIDER
                locationManager.isProviderEnabled(LocationManager.NETWORK_PROVIDER) ->
                    LocationManager.NETWORK_PROVIDER
                else -> {
                    cont.resume(OpsResult.Err(OpsError.business("LOC_OFF", "location provider off")))
                    return@suspendCancellableCoroutine
                }
            }
            val listener = object : LocationListener {
                override fun onLocationChanged(location: Location) {
                    runCatching { locationManager.removeUpdates(this) }
                    if (cont.isActive) cont.resume(OpsResult.Ok(location.toGeo()))
                }

                @Deprecated("Deprecated in Java")
                override fun onStatusChanged(provider: String?, status: Int, extras: Bundle?) = Unit

                override fun onProviderEnabled(provider: String) = Unit
                override fun onProviderDisabled(provider: String) = Unit
            }
            locationManager.requestLocationUpdates(
                provider,
                0L,
                0f,
                listener,
                Looper.getMainLooper(),
            )
            cont.invokeOnCancellation {
                runCatching { locationManager.removeUpdates(listener) }
            }
        }

    private fun Location.toGeo(): GeoPoint = GeoPoint(latitude = latitude, longitude = longitude)
}
