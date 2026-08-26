#!/usr/bin/env python3
"""Import prod RECORD samples into testdata/ fixtures (skip duplicates).

Accepts either:
  {path: [req_json_string, ...]}                         # legacy analyze_prod_log
  {path: [{req: obj, res: obj|null}, ...]}               # extract_record_unique
"""
from __future__ import annotations

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
TESTDATA = os.path.join(ROOT, "testdata")

PATH_TO_DIR = {
    "/device/paas/helmetLock": "device_paas_helmetLock",
    "/device/paas/deviceInfo": "device_paas_deviceInfo",
    "/device/paas/bluetooth": "device_paas_bluetooth",
    "/device/scanLocation/change": "device_scanLocation_change",
    "/device/paas/lock": "device_paas_lock",
    "/device/paas/defend": "device_paas_defend",
    "/device/trajectory/saveDb": "device_trajectory_saveDb",
    "/device/paas/setInnerParam": "device_paas_setInnerParam",
    "/device/paas/voice": "device_paas_voice",
    "/device/paas/replyStopMove": "device_paas_replyStopMove",
    "/device/paas/dashboard": "device_paas_dashboard",
}


def path_to_dir(url: str) -> str:
    return url.strip("/").replace("/", "_")


def normalize_req(obj) -> str:
    return json.dumps(obj, sort_keys=True, ensure_ascii=False)


def existing_reqs(dir_path: str) -> set[str]:
    seen = set()
    if not os.path.isdir(dir_path):
        return seen
    for name in os.listdir(dir_path):
        if not name.endswith(".json"):
            continue
        fp = os.path.join(dir_path, name)
        with open(fp, "r", encoding="utf-8") as f:
            fx = json.load(f)
        req = fx.get("req")
        if req is not None:
            seen.add(normalize_req(req))
    return seen


def next_index(dir_path: str) -> int:
    if not os.path.isdir(dir_path):
        return 0
    nums = []
    for name in os.listdir(dir_path):
        m = re.match(r"^(\d+)\.json$", name)
        if m:
            nums.append(int(m.group(1)))
    return max(nums) + 1 if nums else 0


def iter_samples(entries):
    """Yield (req_obj, res_obj|None) from legacy or new sample formats."""
    for item in entries:
        if isinstance(item, str):
            yield json.loads(item), None
        elif isinstance(item, dict) and "req" in item:
            req = item["req"]
            if isinstance(req, str):
                req = json.loads(req)
            yield req, item.get("res")
        else:
            raise TypeError(f"unsupported sample entry: {type(item)}")


def default_rep():
    return {"success": True, "code": "0", "msg": "成功", "data": None}


def main(samples_path: str) -> None:
    with open(samples_path, "r", encoding="utf-8") as f:
        samples = json.load(f)

    # Tag fixtures with source log basename when possible.
    base = os.path.basename(samples_path)
    source = base.replace("_record_samples.json", "").replace(".json", "")
    clazz = f"prod RECORD sample ({source})"

    added = 0
    skipped = 0
    for url, entries in samples.items():
        dir_name = PATH_TO_DIR.get(url) or path_to_dir(url)
        dir_path = os.path.join(TESTDATA, dir_name)
        os.makedirs(dir_path, exist_ok=True)
        seen = existing_reqs(dir_path)
        idx = next_index(dir_path)

        for req_obj, res_obj in iter_samples(entries):
            key = normalize_req(req_obj)
            if key in seen:
                skipped += 1
                continue
            rep = res_obj if isinstance(res_obj, dict) else default_rep()
            fixture = {
                "url": url,
                "clazz": clazz,
                "cusTime": 0,
                "req": req_obj,
                "rep": rep,
            }
            out = os.path.join(dir_path, f"{idx:03d}.json")
            with open(out, "w", encoding="utf-8") as f:
                json.dump(fixture, f, ensure_ascii=False, indent=2)
                f.write("\n")
            seen.add(key)
            idx += 1
            added += 1
            print(f"  + {out}")

    print(f"\nDone: added={added} skipped={skipped}")


if __name__ == "__main__":
    path = (
        sys.argv[1]
        if len(sys.argv) > 1
        else os.path.join(
            ROOT,
            "prod_ebike-device-paas-go-55b8dcf94c-52nf9_ebike-device-paas-go_record_samples.json",
        )
    )
    main(path)
