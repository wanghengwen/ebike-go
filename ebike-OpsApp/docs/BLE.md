# BLE boundary

Legacy Android `LruBle` and iOS `XCBLETool.framework` are **closed-source**. This open repo cannot ship a working hardware BLE transport or GroupMoveCar / radar cluster unlock.

## Runtime selection

| `features.bleTransport` | Implementation |
|-------------------------|----------------|
| `simulator` (default) | `SimulatorBleTransport` — shared flows without real bikes |
| `native` / `real` | `UnavailableBleTransport` until a private module injects a real `BleTransport` |

Host apps that own the vendor SDK should pass:

```kotlin
OpsApp.create(config = config, bleTransport = YourVendorBleTransport())
```

`VehicleControlPolicy` falls back to network when BLE is unavailable (`BlePreferred`), so field ops still work over the air.

See also `docs/OPEN-SOURCE.md`.
