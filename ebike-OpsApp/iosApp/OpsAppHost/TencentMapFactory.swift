#if OPS_TENCENT_MAP

import CoreLocation
import Foundation
import SharedUi
import UIKit

/// 腾讯地图版的车点图，对应 Android 宿主的 `TencentMapView`。
///
/// 只在真机上编译（`OPS_TENCENT_MAP` 只在 iphoneos 定义）：CocoaPods 拉下来的
/// `QMapKit.framework` 是老式 fat framework（armv7 / i386 / x86_64 / arm64，
/// 其中 arm64 是设备 slice），没有 arm64 模拟器 slice，Apple Silicon 的模拟器链不上。
/// 模拟器构建时这个文件被整体排除，共享层自动落到 `MapKitOpsMapRenderer`。
/// SDK 类型经 bridging header 引入，所以这里没有 `import QMapKit`。
///
/// 聚合在宿主按 zoom 网格计算（对齐 Android `MapClusterer.cellDegreesForZoom`）。
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
        // 散点跟车点在腾讯这边是同一套 annotation，只是配色不同（abnormal 走红色）。
        TencentMapHostView(onSelectCarId: { _ in }, onSelectCluster: { _ in })
    }
}

private final class TencentMapHostView: NSObject, IosHostMapView, QMapViewDelegate {

    private let mapView = QMapView(frame: .zero)
    private let onSelectCarId: (String) -> Void
    private let onSelectCluster: ([String]) -> Void

    /// QMapView 是命令式的，Compose 每次重组都会调 apply；记住上次输入才不会
    /// 把相机反复重置（用户会拖不动地图）。
    private var sourcePins: [IosHostMapPin] = []
    private var wantsCluster = true
    private var lastClusterZoom: Double = -1
    private var pinsIdentity: String?
    private var lastWantsCluster: Bool?
    private var overlayKey: String?
    private var selectedCarId: String?
    /// 跟 Android / MapKit 一样：按「坐标集 + 围栏 + 轨迹 + nonce」贴视野，
    /// 不能只盯 fitNonce——车点是异步到的，首帧 nonce=0 且 pins 为空，
    /// 之后 pins 来了 nonce 仍是 0，相机就会永远停在默认比例。
    private var cameraFittedFor: String?
    private var cameraSelectedCarId: String?
    private var pendingFitRegion: QCoordinateRegion?
    // Kotlin 的 Int 导出到 Swift 是 Int32，别写成 Int，否则每处都要转一次。
    private var zoomInNonce: Int32 = -1
    private var zoomOutNonce: Int32 = -1
    private var followNonce: Int32 = -1
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
        }
        let key = parts.joined(separator: "|")
        sourcePins = update.pins
        wantsCluster = update.clusterOverview
        let identityChanged = pinsIdentity != key || lastWantsCluster != wantsCluster
        pinsIdentity = key
        lastWantsCluster = wantsCluster
        if identityChanged {
            refreshClusteredAnnotations(force: true)
        }
        if selectedCarId != update.selectedCarId {
            selectedCarId = update.selectedCarId
            if let target = update.selectedCarId,
               let hit = mapView.annotations
                   .compactMap({ $0 as? OpsPointAnnotation })
                   .first(where: { $0.pin.id == target || $0.pin.memberIds.contains(target) }) {
                mapView.selectAnnotation(hit, animated: true)
            }
        }
    }

    private func refreshClusteredAnnotations(force: Bool) {
        let zoom = mapView.zoomLevel
        if !force && abs(zoom - lastClusterZoom) < 0.08 { return }
        lastClusterZoom = zoom
        let lat = sourcePins.first(where: { $0.lat != 0 || $0.lng != 0 })?.lat ?? 30
        let display = wantsCluster
            ? Self.clustered(sourcePins, zoom: zoom, latitude: lat)
            : sourcePins
        mapView.removeAnnotations(mapView.annotations)
        mapView.addAnnotations(display.map { OpsPointAnnotation(pin: $0) })
        if let target = selectedCarId,
           let hit = mapView.annotations
               .compactMap({ $0 as? OpsPointAnnotation })
               .first(where: { $0.pin.id == target || $0.pin.memberIds.contains(target) }) {
            mapView.selectAnnotation(hit, animated: false)
        }
    }

    private static func clustered(_ pins: [IosHostMapPin], zoom: Double, latitude: Double) -> [IosHostMapPin] {
        guard pins.count > 1 else { return pins }
        let cell = cellDegrees(zoom: zoom, latitude: latitude)
        var buckets: [String: [IosHostMapPin]] = [:]
        for pin in pins {
            let keyLat = Int64(floor(pin.lat / cell))
            let keyLng = Int64(floor(pin.lng / cell))
            let key = "\(keyLat):\(keyLng)"
            buckets[key, default: []].append(pin)
        }
        var out: [IosHostMapPin] = []
        for group in buckets.values {
            if group.count == 1 {
                out.append(group[0])
            } else if let last = group.last {
                let ids = group.map(\.id)
                out.append(
                    IosHostMapPin(
                        id: "cluster:" + ids.sorted().joined(separator: ","),
                        lat: last.lat,
                        lng: last.lng,
                        title: "\(group.count)",
                        subtitle: "vehicles",
                        restBattery: last.restBattery,
                        ridingState: last.ridingState,
                        memberCount: Int32(group.count),
                        memberIds: ids,
                        abnormal: false,
                        iconName: last.iconName
                    )
                )
            }
        }
        if out.count > 100 { return Array(out.prefix(100)) }
        return out
    }

    private static func cellDegrees(zoom: Double, latitude: Double) -> Double {
        let dp: Double
        if zoom < 10 { dp = 60 }
        else if zoom < 13 { dp = 50 }
        else if zoom < 14 { dp = 40 }
        else { dp = 60 }
        let px = dp * 3
        let z = min(max(zoom, 3), 22)
        let metersPerPixel = 156543.03392 * max(abs(cos(latitude * .pi / 180)), 0.2) / pow(2, z)
        return max((px * metersPerPixel) / 111320.0, 1e-6)
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
        let fitKey = fitParts.joined(separator: "|")

        if !valid.isEmpty && fitKey != cameraFittedFor {
            cameraFittedFor = fitKey
            let region: QCoordinateRegion?
            if valid.count == 1 && update.fences.isEmpty && update.trackPoints.isEmpty {
                let only = valid[0]
                region = QCoordinateRegionMake(
                    CLLocationCoordinate2D(latitude: only.lat, longitude: only.lng),
                    QCoordinateSpanMake(Self.singlePinSpan, Self.singlePinSpan)
                )
            } else {
                var coords = valid.map { ($0.lat, $0.lng) }
                coords += update.trackPoints.map { ($0.lat, $0.lng) }
                for fence in update.fences {
                    coords += fence.points.map { ($0.lat, $0.lng) }
                }
                region = Self.regionFor(coords)
            }
            if let region {
                // Compose 首帧 UIKitView 经常还是 0×0（见 boot 日志 drawableSize=0），
                // 这时 setRegion 会被吞掉；先记下来，等有尺寸再补一次。
                applyRegionWhenReady(region, animated: false)
            }
        } else if cameraSelectedCarId != update.selectedCarId,
                  let target = update.selectedCarId,
                  let selected = valid.first(where: {
                      $0.memberCount <= 1 && ($0.id == target || $0.memberIds.contains(target))
                  }) {
            // 只在选中车变化时跟过去，避免每帧 setRegion 把用户拖动手势顶掉。
            cameraSelectedCarId = target
            let region = QCoordinateRegionMake(
                CLLocationCoordinate2D(latitude: selected.lat, longitude: selected.lng),
                QCoordinateSpanMake(Self.selectedSpan, Self.selectedSpan)
            )
            applyRegionWhenReady(region, animated: true)
        } else if update.selectedCarId == nil {
            cameraSelectedCarId = nil
        }

        // 尺寸从 0 变成有效后，把挂起的视野补上。
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
                mapView.centerCoordinate = CLLocationCoordinate2D(
                    latitude: update.followLat,
                    longitude: update.followLng
                )
            }
        }
    }

    private func applyRegionWhenReady(_ region: QCoordinateRegion, animated: Bool) {
        if mapView.bounds.width > 1 && mapView.bounds.height > 1 {
            pendingFitRegion = nil
            mapView.setRegion(region, animated: animated)
        } else {
            pendingFitRegion = region
            // 下一 runloop 再试一次：UIKitView 布局往往就在这一拍之后完成。
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

    private static let singlePinSpan: CLLocationDegrees = 0.012
    private static let selectedSpan: CLLocationDegrees = 0.006

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
            // 1.4 倍留白，跟 Android 的 fitBounds 边距观感对齐。
            QCoordinateSpanMake(
                max((maxLat - minLat) * 1.4, 0.004),
                max((maxLng - minLng) * 1.4, 0.004)
            )
        )
    }

    // MARK: QMapViewDelegate

    func mapView(_ mapView: QMapView!, viewFor annotation: QAnnotation!) -> QAnnotationView! {
        guard let point = annotation as? OpsPointAnnotation else { return nil }
        let reuseId = point.pin.memberCount > 1 ? "ops-cluster" : "ops-vehicle"
        let view = mapView.dequeueReusableAnnotationView(withIdentifier: reuseId)
            ?? QAnnotationView(annotation: annotation, reuseIdentifier: reuseId)
        view?.annotation = annotation
        view?.canShowCallout = false
        view?.centerOffset = .zero
        if point.pin.memberCount > 1 {
            view?.image = Self.clusterImage(count: Int(point.pin.memberCount))
        } else if let image = UIImage(named: point.pin.iconName)
            ?? UIImage(named: "\(point.pin.iconName).webp") {
            view?.image = image
        } else {
            view?.image = UIImage(named: "icon_vehicle") ?? UIImage(named: "icon_vehicle.webp")
        }
        return view
    }

    private static func clusterImage(count: Int) -> UIImage {
        let size: CGFloat = 20
        let renderer = UIGraphicsImageRenderer(size: CGSize(width: size, height: size))
        return renderer.image { ctx in
            UIColor(red: 41 / 255, green: 95 / 255, blue: 204 / 255, alpha: 1).setFill()
            ctx.cgContext.fillEllipse(in: CGRect(x: 0, y: 0, width: size, height: size))
            let text = count > 10 ? "\(count)" : " \(count) "
            let attrs: [NSAttributedString.Key: Any] = [
                .font: UIFont.boldSystemFont(ofSize: 11),
                .foregroundColor: UIColor.white,
            ]
            let ns = text as NSString
            let textSize = ns.size(withAttributes: attrs)
            ns.draw(
                at: CGPoint(x: (size - textSize.width) / 2, y: (size - textSize.height) / 2),
                withAttributes: attrs
            )
        }
    }

    func mapView(_ mapView: QMapView!, regionDidChangeAnimated animated: Bool) {
        refreshClusteredAnnotations(force: false)
    }

    func mapView(_ mapView: QMapView!, didSelect view: QAnnotationView!) {
        guard let point = view.annotation as? OpsPointAnnotation else { return }
        let pin = point.pin
        if pin.memberCount > 1 {
            // Legacy handleClusterClick: zoom+1 under 16 (center on cluster), else vehicle list.
            if mapView.zoomLevel < 16 {
                mapView.deselectAnnotation(view.annotation, animated: false)
                let next = min(mapView.zoomLevel + 1, mapView.maxZoomLevel)
                mapView.setCenterCoordinate(
                    CLLocationCoordinate2D(latitude: pin.lat, longitude: pin.lng),
                    animated: true
                )
                mapView.setZoomLevel(next, animated: true)
            } else {
                onSelectCluster(pin.memberIds.isEmpty ? [pin.id] : pin.memberIds)
            }
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
        // Key 没有绑定这个 bundleId 时会走到这里。不静默：地图空白很难查。
        NSLog("[OpsApp] Tencent map failed to load: %@", error?.localizedDescription ?? "unknown")
    }
}

private extension QMapView {
    /// `overlays` 是无泛型的 `NSArray`，直接喂给 `removeOverlays(_:)`（要 `[QOverlay]`）
    /// 会让 Swift 编译器报 "failed to produce diagnostic"，所以这里显式收窄一次。
    func removeAllOverlays() {
        removeOverlays(overlays.compactMap { $0 as? QOverlay })
    }
}

private final class OpsPointAnnotation: QPointAnnotation {
    let pin: IosHostMapPin

    init(pin: IosHostMapPin) {
        self.pin = pin
        super.init()
        coordinate = CLLocationCoordinate2D(latitude: pin.lat, longitude: pin.lng)
        title = pin.title
        subtitle = pin.subtitle
    }
}

#endif
