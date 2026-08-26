#!/usr/bin/env python3
"""Analyze prod ebike-device-paas-go log for RECORD/SHADOW/ACCESS stats."""
import json
import re
import sys
from collections import Counter, defaultdict

def diff_obj(a, b, prefix=""):
    diffs = []
    if type(a) != type(b):
        diffs.append(f"{prefix}: type {type(a).__name__} vs {type(b).__name__}")
        return diffs
    if isinstance(a, dict):
        keys = set(a.keys()) | set(b.keys())
        for k in sorted(keys):
            p = f"{prefix}.{k}" if prefix else k
            if k not in a:
                diffs.append(f"{p}: missing in Java")
            elif k not in b:
                diffs.append(f"{p}: missing in Go")
            else:
                diffs.extend(diff_obj(a[k], b[k], p))
    elif isinstance(a, list):
        if len(a) != len(b):
            diffs.append(f"{prefix}: len {len(a)} vs {len(b)}")
        for idx, (x, y) in enumerate(zip(a, b)):
            diffs.extend(diff_obj(x, y, f"{prefix}[{idx}]"))
    elif a != b:
        diffs.append(f"{prefix}: {repr(a)[:100]} vs {repr(b)[:100]}")
    return diffs


def main(log_path):
    with open(log_path, "r", encoding="utf-8", errors="replace") as f:
        lines = f.readlines()

    print("Total lines:", len(lines))

    record_paths = Counter()
    record_failures = []
    access_paths = Counter()
    shadow_match = shadow_diff = 0
    errors = []
    record_samples = defaultdict(list)
    record_req_only = defaultdict(list)

    i = 0
    while i < len(lines):
        line = lines[i].strip()
        if "[RECORD]" in line:
            m = re.search(r"Path: (\S+)", line)
            path = m.group(1) if m else "unknown"
            record_paths[path] += 1
            req = res = None
            if i + 1 < len(lines) and lines[i + 1].startswith("Req:"):
                req = lines[i + 1][4:].strip()
            if i + 2 < len(lines) and lines[i + 2].startswith("Res:"):
                res = lines[i + 2][4:].strip()
            if req and len(record_req_only[path]) < 3:
                record_req_only[path].append(req)
            if req and len(record_samples[path]) < 2:
                record_samples[path].append((req, res))
            if res and '"success":false' in res:
                record_failures.append((path, req, res))
        elif "[SHADOW MATCH]" in line:
            shadow_match += 1
        elif "[SHADOW DIFF]" in line:
            shadow_diff += 1
            diff_java = diff_go = None
            if i + 2 < len(lines) and lines[i + 2].startswith("Java:"):
                diff_java = lines[i + 2][5:].strip()
            if i + 3 < len(lines) and lines[i + 3].startswith("Go:"):
                diff_go = lines[i + 3][3:].strip()
            if diff_java and diff_go:
                try:
                    j = json.loads(diff_java)
                    g = json.loads(diff_go)
                    diffs = diff_obj(j, g)
                    print("\n=== SHADOW DIFF field differences ===")
                    for d in diffs[:40]:
                        print(" ", d)
                except json.JSONDecodeError as e:
                    print("SHADOW DIFF JSON error:", e)
                    for pos, (a, b) in enumerate(zip(diff_java, diff_go)):
                        if a != b:
                            print("First char diff at", pos)
                            print(" Java:", diff_java[max(0, pos - 80) : pos + 80])
                            print(" Go:  ", diff_go[max(0, pos - 80) : pos + 80])
                            break
        elif "[ACCESS]" in line and "POST" in line:
            m = re.search(r"POST (\S+)", line)
            if m:
                access_paths[m.group(1)] += 1
            if i + 2 < len(lines) and lines[i + 2].startswith("Res:"):
                res = lines[i + 2][4:].strip()
                if '"success":false' in res:
                    try:
                        r = json.loads(res)
                        errors.append((r.get("code"), r.get("msg"), line.strip()[:100]))
                    except json.JSONDecodeError:
                        pass
        i += 1

    print("\n=== RECORD paths ===")
    for p, c in record_paths.most_common():
        print(f"  {p}: {c}")
    print("RECORD total:", sum(record_paths.values()))
    print("RECORD failures:", len(record_failures))
    for path, req, res in record_failures:
        print(f"\n  FAIL path={path}")
        print(f"  Req: {req[:300]}...")
        print(f"  Res: {res[:400]}")

    print("\n=== SHADOW ===")
    print("MATCH:", shadow_match, "DIFF:", shadow_diff)

    print("\n=== ACCESS top 20 ===")
    for p, c in access_paths.most_common(20):
        print(f"  {p}: {c}")
    print("ACCESS total:", sum(access_paths.values()))

    print("\n=== Business failures (Res success:false) ===")
    err_codes = Counter((c, m) for c, m, _ in errors)
    for (c, m), cnt in err_codes.most_common():
        print(f"  {c} ({cnt}x): {m}")

    print("\n=== RECORD unique request samples (for testdata) ===")
    for p in sorted(record_req_only.keys()):
        print(f"\n--- {p} ---")
        for idx, req in enumerate(record_req_only[p]):
            try:
                obj = json.loads(req)
                print(f"  sample{idx+1} keys:", sorted(obj.keys()))
                cc = obj.get("commandContext", {})
                print(f"    commandContext keys:", sorted(cc.keys()) if cc else None)
            except json.JSONDecodeError:
                print(f"  sample{idx+1}: (invalid json) {req[:200]}")

    # Export samples to json file
    out = {}
    for p, reqs in record_req_only.items():
        out[p] = reqs
    out_path = log_path.replace(".log", "_record_samples.json")
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(out, f, ensure_ascii=False, indent=2)
    print(f"\nExported RECORD samples to {out_path}")


if __name__ == "__main__":
    main(sys.argv[1] if len(sys.argv) > 1 else r"prod_ebike-device-paas-go-dfdbf8cbd-j444z_ebike-device-paas-go.log")
