#!/usr/bin/env python3
"""Extract unique RECORD (req,res) samples from a prod paas-go log.

Dedupes by normalized request JSON. Writes:
  <log>_record_samples.json  — {path: [{req, res}, ...]}
"""
from __future__ import annotations

import json
import re
import sys
from collections import defaultdict
from pathlib import Path

# Cap unique shapes kept per path (enough for bind / shape coverage).
MAX_PER_PATH = 15


def normalize_req(obj) -> str:
    return json.dumps(obj, sort_keys=True, ensure_ascii=False)


def main(log_path: str) -> None:
    path = Path(log_path)
    text = path.read_text(encoding="utf-8", errors="replace")
    lines = text.splitlines()

    # path -> list of (req_obj, res_obj|None)
    by_path: dict[str, list[tuple[dict, object]]] = defaultdict(list)
    seen: dict[str, set[str]] = defaultdict(set)
    totals = defaultdict(int)
    parse_fail = 0

    i = 0
    while i < len(lines):
        line = lines[i]
        if "[RECORD]" not in line:
            i += 1
            continue
        m = re.search(r"Path: (\S+)", line)
        if not m:
            i += 1
            continue
        api = m.group(1)
        totals[api] += 1

        req_s = res_s = None
        if i + 1 < len(lines) and lines[i + 1].startswith("Req:"):
            req_s = lines[i + 1][4:].strip()
        if i + 2 < len(lines) and lines[i + 2].startswith("Res:"):
            res_s = lines[i + 2][4:].strip()

        if not req_s:
            i += 1
            continue
        try:
            req_obj = json.loads(req_s)
        except json.JSONDecodeError:
            parse_fail += 1
            i += 1
            continue

        key = normalize_req(req_obj)
        if key in seen[api]:
            i += 3 if res_s else 2
            continue
        if len(by_path[api]) >= MAX_PER_PATH:
            i += 3 if res_s else 2
            continue

        res_obj = None
        if res_s:
            try:
                res_obj = json.loads(res_s)
            except json.JSONDecodeError:
                res_obj = None

        seen[api].add(key)
        by_path[api].append((req_obj, res_obj))
        i += 3 if res_s else 2

    out = {
        api: [{"req": req, "res": res} for req, res in samples]
        for api, samples in sorted(by_path.items())
    }
    out_path = str(path).replace(".log", "_record_samples.json")
    Path(out_path).write_text(
        json.dumps(out, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )

    print(f"log: {path.name}")
    print(f"RECORD total: {sum(totals.values())}  parse_fail: {parse_fail}")
    print(f"unique exported: {sum(len(v) for v in out.values())} -> {out_path}")
    for api, c in sorted(totals.items(), key=lambda x: -x[1]):
        print(f"  {api}: total={c} unique_kept={len(out.get(api, []))}")


if __name__ == "__main__":
    main(
        sys.argv[1]
        if len(sys.argv) > 1
        else str(
            Path(__file__).resolve().parents[1]
            / "prod_ebike-device-paas-go-55b8dcf94c-52nf9_ebike-device-paas-go.log"
        )
    )
