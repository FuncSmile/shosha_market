// Minimal shims for Web Bluetooth types used in @vueuse/core types
// These are intentionally permissive (any) to avoid blocking builds when
// the project doesn't require full Web Bluetooth typings.

declare type BluetoothLEScanFilter = any;
declare type BluetoothServiceUUID = any;

interface BluetoothDevice {
  // permissive shape
  id?: string;
  name?: string | null;
  gatt?: BluetoothRemoteGATTServer | undefined;
}

interface BluetoothRemoteGATTServer {
  connected?: boolean;
  // methods are intentionally left untyped
}
