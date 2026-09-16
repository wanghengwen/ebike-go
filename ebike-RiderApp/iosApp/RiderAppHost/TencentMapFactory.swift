#if RIDER_TENCENT_MAP

import CoreLocation
import Foundation
import SharedUi
import UIKit

/// 腾讯地图版的车点图，对应 Android 宿主的 `TencentMapView`。
///
/// 只在真机上编译（`RIDER_TENCENT_MAP` 只在 iphoneos 定义）：CocoaPods 拉下来的
/// `QMapKit.framework` 是老式 fat framework，没有 arm64 模拟器 slice。
/// SDK 类型经 bridging header 引入，所以这里没有 `import QMapKit`。
final class TencentMapFactory: NSObject, IosHostMapFactory {

    private static var registered = false

    static func register(apiKey: String) {
        if !registered {
            QMapServices.shared().apiKey = apiKey
            registered = true
        }
        IosMapHost.shared.factory = TencentMapFactory()
    }

    func isReady() -> Bool { TencentMapFactory.registered }

    func createPinsView(
        onSelectCarId: @escaping (String) -> Void,
        onSelectCluster: @escaping ([String]) -> Void
    ) -> IosHostMapView {
        TencentMapHostView(onSelectCarId: onSelectCarId, onSelectCluster: onSelectCluster)
    }

    func createScatterView() -> IosHostMapView? {
        TencentMapHostView(onSelectCarId: { _ in }, onSelectCluster: { _ in })
    }
}

private final class TencentMapHostView: NSObject, IosHostMapView, QMapViewDelegate {

    private let mapView = QMapView(frame: .zero)
    private let onSelectCarId: (String) -> Void
    private let onSelectCluster: ([String]) -> Void

    private var pinsKey: String?
    private var overlayKey: String?
    private var selectedCarId: String?
    private var cameraFittedFor: String?
    private var cameraSelectedCarId: String?
    private var pendingFitRegion: QCoordinateRegion?
    // 与 RiderMapSpec 默认 0 对齐；初值 -1 会在首帧被当成用户点了放大/缩小。
    private var zoomInNonce: Int32 = 0
    private var zoomOutNonce: Int32 = 0
    private var followNonce: Int32 = 0
    private var satellite: Bool?

    init(onSelectCarId: @escaping (String) -> Void, onSelectCluster: @escaping ([String]) -> Void) {
        self.onSelectCarId = onSelectCarId
        self.onSelectCluster = onSelectCluster
        super.init()
        mapView.delegate = self
        mapView.showsScale = true
        mapView.showsCompass = true
    }

    func view() -> UIView { mapView }

    func dispose() {
        mapView.delegate = nil
        mapView.removeAnnotations(mapView.annotations)
        mapView.removeAllOverlays()
    }

    func apply(update: IosHostMapUpdate) {
        applyMapType(update.satellite)
        applyPins(update)
        applyOverlays(update)
        applyCamera(update)
    }

    private func applyMapType(_ wantSatellite: Bool) {
        guard satellite != wantSatellite else { return }
        satellite = wantSatellite
        mapView.mapType = wantSatellite ? .satellite : .standard
    }

    private func applyPins(_ update: IosHostMapUpdate) {
        var parts: [String] = []
        for pin in update.pins {
            parts.append(pin.id)
            parts.append(String(pin.lat))
            parts.append(String(pin.lng))
            parts.append(String(pin.memberCount))
        }
        let key = parts.joined(separator: "|")
        if pinsKey != key {
            pinsKey = key
            mapView.removeAnnotations(mapView.annotations)
            mapView.addAnnotations(update.pins.map { RiderPointAnnotation(pin: $0) })
        }
        if selectedCarId != update.selectedCarId {
            selectedCarId = update.selectedCarId
            if let target = update.selectedCarId,
               let hit = mapView.annotations
                   .compactMap({ $0 as? RiderPointAnnotation })
                   .first(where: { $0.pin.id == target || $0.pin.memberIds.contains(target) }) {
                mapView.selectAnnotation(hit, animated: true)
            }
        }
    }

    private func applyOverlays(_ update: IosHostMapUpdate) {
        var parts: [String] = [String(update.trackPoints.count)]
        for fence in update.fences {
            parts.append(fence.id)
            parts.append(String(fence.points.count))
        }
        let key = parts.joined(separator: "|")
        guard overlayKey != key else { return }
        overlayKey = key
        mapView.removeAllOverlays()
        for fence in update.fences where fence.points.count >= 3 {
            var coords = fence.points.map {
                CLLocationCoordinate2D(latitude: $0.lat, longitude: $0.lng)
            }
            let polygon = QPolygon(coordinates: &coords, count: UInt(coords.count))
            mapView.add(polygon)
        }
        if update.trackPoints.count >= 2 {
            var coords = update.trackPoints.map {
                CLLocationCoordinate2D(latitude: $0.lat, longitude: $0.lng)
            }
            let line = QPolyline(coordinates: &coords, count: UInt(coords.count))
            mapView.add(line)
        }
    }

    private func applyCamera(_ update: IosHostMapUpdate) {
        let valid = update.pins.filter { $0.lat != 0 || $0.lng != 0 }
        var fitParts: [String] = []
        for pin in valid {
            fitParts.append("\(pin.id):\(pin.lat),\(pin.lng)")
        }
        fitParts.append("f\(update.fences.count)")
        fitParts.append("t\(update.trackPoints.count)")
        fitParts.append("n\(update.fitNonce)")
        fitParts.append(update.clusterOverview ? "home" : "ride")
        if update.hasFollowTarget {
            fitParts.append("c\(update.followLat),\(update.followLng)")
        }
        let fitKey = fitParts.joined(separator: "|")

        if !valid.isEmpty && fitKey != cameraFittedFor {
            cameraFittedFor = fitKey
            if update.clusterOverview {
                // 首页对齐 UniApp `scale=17`：以定位点（或车点中心）定焦，不把附近车辆全塞进视野。
                let center: CLLocationCoordinate2D
                if update.hasFollowTarget {
                    center = CLLocationCoordinate2D(
                        latitude: update.followLat,
                        longitude: update.followLng
                    )
                } else {
                    let lat = valid.map(\.lat).reduce(0, +) / Double(valid.count)
                    let lng = valid.map(\.lng).reduce(0, +) / Double(valid.count)
                    center = CLLocationCoordinate2D(latitude: lat, longitude: lng)
                }
                applyCenterZoomWhenReady(center, zoom: Self.homeZoomLevel, animated: false)
            } else if valid.count == 1 && update.fences.isEmpty && update.trackPoints.isEmpty {
                let only = valid[0]
                applyCenterZoomWhenReady(
                    CLLocationCoordinate2D(latitude: only.lat, longitude: only.lng),
                    zoom: Self.singlePinZoomLevel,
                    animated: false
                )
            } else {
                var coords = valid.map { ($0.lat, $0.lng) }
                coords += update.trackPoints.map { ($0.lat, $0.lng) }
                for fence in update.fences {
                    coords += fence.points.map { ($0.lat, $0.lng) }
                }
                if let region = Self.regionFor(coords) {
                    applyRegionWhenReady(region, animated: false)
                }
            }
        } else if cameraSelectedCarId != update.selectedCarId,
                  let target = update.selectedCarId,
                  let selected = valid.first(where: {
                      $0.memberCount <= 1 && ($0.id == target || $0.memberIds.contains(target))
                  }) {
            cameraSelectedCarId = target
            applyCenterZoomWhenReady(
                CLLocationCoordinate2D(latitude: selected.lat, longitude: selected.lng),
                zoom: Self.selectedZoomLevel,
                animated: true
            )
        } else if update.selectedCarId == nil {
            cameraSelectedCarId = nil
        }

        flushPendingFitIfPossible()

        if zoomInNonce != update.zoomInNonce {
            zoomInNonce = update.zoomInNonce
            mapView.setZoomLevel(min(mapView.zoomLevel + 1, mapView.maxZoomLevel), animated: true)
        }
        if zoomOutNonce != update.zoomOutNonce {
            zoomOutNonce = update.zoomOutNonce
            mapView.setZoomLevel(max(mapView.zoomLevel - 1, mapView.minZoomLevel), animated: true)
        }
        if followNonce != update.followNonce {
            followNonce = update.followNonce
            if update.hasFollowTarget {
                applyCenterZoomWhenReady(
                    CLLocationCoordinate2D(
                        latitude: update.followLat,
                        longitude: update.followLng
                    ),
                    zoom: Self.homeZoomLevel,
                    animated: true
                )
            }
        }
    }

    private func applyCenterZoomWhenReady(
        _ center: CLLocationCoordinate2D,
        zoom: CGFloat,
        animated: Bool
    ) {
        let apply = { [weak self] in
            guard let self else { return }
            self.mapView.centerCoordinate = center
            self.mapView.setZoomLevel(zoom, animated: animated)
        }
        if mapView.bounds.width > 1 && mapView.bounds.height > 1 {
            pendingFitRegion = nil
            apply()
        } else {
            // 尺寸未就绪时先记 region 兜底；布局完成后再补 zoom。
            pendingFitRegion = QCoordinateRegionMake(
                center,
                QCoordinateSpanMake(Self.homeSpan, Self.homeSpan)
            )
            DispatchQueue.main.async { [weak self] in
                self?.flushPendingFitIfPossible()
                if self?.mapView.bounds.width ?? 0 > 1 {
                    apply()
                }
            }
        }
    }

    private func applyRegionWhenReady(_ region: QCoordinateRegion, animated: Bool) {
        if mapView.bounds.width > 1 && mapView.bounds.height > 1 {
            pendingFitRegion = nil
            mapView.setRegion(region, animated: animated)
        } else {
            pendingFitRegion = region
            DispatchQueue.main.async { [weak self] in
                self?.flushPendingFitIfPossible()
            }
        }
    }

    private func flushPendingFitIfPossible() {
        guard let region = pendingFitRegion else { return }
        guard mapView.bounds.width > 1 && mapView.bounds.height > 1 else { return }
        pendingFitRegion = nil
        mapView.setRegion(region, animated: false)
    }

    /// UniApp 首页 / 骑行 `scale=17`。
    private static let homeZoomLevel: CGFloat = 17
    private static let singlePinZoomLevel: CGFloat = 16
    private static let selectedZoomLevel: CGFloat = 17
    private static let homeSpan: CLLocationDegrees = 0.003

    private static func regionFor(_ coords: [(Double, Double)]) -> QCoordinateRegion? {
        let valid = coords.filter { $0.0 != 0 || $0.1 != 0 }
        guard !valid.isEmpty else { return nil }
        let lats = valid.map(\.0)
        let lngs = valid.map(\.1)
        let minLat = lats.min()!, maxLat = lats.max()!
        let minLng = lngs.min()!, maxLng = lngs.max()!
        return QCoordinateRegionMake(
            CLLocationCoordinate2D(
                latitude: (minLat + maxLat) / 2,
                longitude: (minLng + maxLng) / 2
            ),
            QCoordinateSpanMake(
                max((maxLat - minLat) * 1.4, 0.004),
                max((maxLng - minLng) * 1.4, 0.004)
            )
        )
    }

    // MARK: QMapViewDelegate

    func mapView(_ mapView: QMapView!, viewFor annotation: QAnnotation!) -> QAnnotationView! {
        guard let point = annotation as? RiderPointAnnotation else { return nil }
        let reuseId = "rider-pin"
        let reused = mapView.dequeueReusableAnnotationView(withIdentifier: reuseId)
        guard let view = reused
            ?? QPinAnnotationView(annotation: annotation, reuseIdentifier: reuseId) else {
            return nil
        }
        view.annotation = annotation
        view.canShowCallout = !point.pin.title.isEmpty
        if let pinView = view as? QPinAnnotationView {
            pinView.pinColor = point.pin.markerColor
        }
        return view
    }

    func mapView(_ mapView: QMapView!, didSelect view: QAnnotationView!) {
        guard let point = view.annotation as? RiderPointAnnotation else { return }
        let pin = point.pin
        if pin.memberCount > 1 {
            onSelectCluster(pin.memberIds.isEmpty ? [pin.id] : pin.memberIds)
        } else {
            onSelectCarId(pin.id)
        }
    }

    func mapView(_ mapView: QMapView!, viewFor overlay: QOverlay!) -> QOverlayView! {
        if let polygon = overlay as? QPolygon {
            let renderer = QPolygonView(polygon: polygon)
            renderer?.strokeColor = .systemBlue
            renderer?.lineWidth = 2
            renderer?.fillColor = UIColor.systemBlue.withAlphaComponent(0.12)
            return renderer
        }
        if let line = overlay as? QPolyline {
            let renderer = QPolylineView(polyline: line)
            renderer?.strokeColor = .systemOrange
            renderer?.lineWidth = 4
            return renderer
        }
        return nil
    }

    func mapViewDidFailLoadingMap(_ mapView: QMapView!, withError error: Error!) {
        NSLog("[RiderApp] Tencent map failed to load: %@", error?.localizedDescription ?? "unknown")
    }
}

private extension QMapView {
    func removeAllOverlays() {
        removeOverlays(overlays.compactMap { $0 as? QOverlay })
    }
}

private final class RiderPointAnnotation: QPointAnnotation {
    let pin: IosHostMapPin

    init(pin: IosHostMapPin) {
        self.pin = pin
        super.init()
        coordinate = CLLocationCoordinate2D(latitude: pin.lat, longitude: pin.lng)
        title = pin.title
        subtitle = pin.subtitle
    }
}

private extension IosHostMapPin {
    var markerColor: QPinAnnotationColor {
        if memberCount > 1 { return QPinAnnotationColorPurple }
        if abnormal || (restBattery >= 1 && restBattery <= 30) { return QPinAnnotationColorRed }
        return QPinAnnotationColorGreen
    }
}

#endif
