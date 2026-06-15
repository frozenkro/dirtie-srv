#!/usr/bin/env python3
"""
Smoke tests for dirtie-srv endpoints on the k3s deployment.

Usage:
    python3 smoke_test.py [--url BASE_URL]

Default target is http://rpic1:8080.
Records a unique smoke-test user to avoid collisions.
"""

import argparse
import json
import sys
import time
import urllib.error
import urllib.request


def parse_args():
    parser = argparse.ArgumentParser(description="Smoke test dirtie-srv endpoints")
    parser.add_argument("--url", default="http://rpic1:8080", help="Base URL of the deployed service")
    parser.add_argument("-v", "--verbose", action="store_true", help="Print response bodies")
    return parser.parse_args()


class SmokeTestRunner:
    def __init__(self, base_url: str, verbose: bool = False):
        self.base_url = base_url.rstrip("/")
        self.verbose = verbose
        self.cookie = None
        self.user_email = f"smoke-{int(time.time())}@example.com"
        self.password = "SmokeP@ss123"
        self.passed = 0
        self.failed = 0

    def log_pass(self, msg):
        print(f"[PASS] {msg}")
        self.passed += 1

    def log_fail(self, msg):
        print(f"[FAIL] {msg}", file=sys.stderr)
        self.failed += 1

    def req(self, path: str, method="GET", data=None, headers=None, with_cookie=False, expect_codes=None):
        url = f"{self.base_url}/{path.lstrip('/')}"
        req_headers = {"Content-Type": "application/json"}
        if headers:
            req_headers.update(headers)

        req = urllib.request.Request(url, method=method, headers=req_headers)
        if data is not None:
            req.data = json.dumps(data).encode("utf-8")

        if with_cookie and self.cookie:
            req.add_header("Cookie", f"dirtie.auth={self.cookie}")

        try:
            with urllib.request.urlopen(req) as resp:
                body = resp.read().decode("utf-8", errors="replace")
                if self.verbose:
                    print(f"  --> HTTP {resp.status}")
                    print(f"  --> Body: {body[:500]}")
                return resp.status, body, resp.headers
        except urllib.error.HTTPError as e:
            body = e.read().decode("utf-8", errors="replace") if e.read() else ""
            if self.verbose:
                print(f"  --> HTTP {e.code}")
                print(f"  --> Body: {body[:500]}")
            return e.code, body, e.headers
        except urllib.error.URLError as e:
            self.log_fail(f"Connection error: {e.reason}")
            return None, "", {}

    def run(self):
        print(f"Testing {self.base_url}")
        print(f"Using test user: {self.user_email}")
        print()

        # GET /
        print("--- GET / ---")
        status, _, _ = self.req("/")
        if status == 200:
            self.log_pass("GET / => 200")
        else:
            self.log_fail(f"GET / expected 200, got {status}")

        # POST /users
        print("\n--- POST /users ---")
        status, _, _ = self.req("/users", method="POST", data={
            "email": self.user_email,
            "password": self.password,
            "name": "Smoke Test",
        })
        if status == 200:
            self.log_pass("POST /users => 200")
        else:
            self.log_fail(f"POST /users expected 200, got {status}")

        # POST /login
        print("\n--- POST /login ---")
        status, _, headers = self.req("/login", method="POST", data={
            "email": self.user_email,
            "password": self.password,
        })
        if status == 200:
            self.log_pass("POST /login => 200")
            # capture cookie
            cookies = headers.get("Set-Cookie", "")
            if cookies:
                for c in cookies.split(","):
                    if "dirtie.auth" in c:
                        self.cookie = c.split(";")[0].split("=")[1].strip()
                        break
            if self.cookie:
                print(f"  Captured auth cookie")
            else:
                self.log_fail("No dirtie.auth cookie after login")
        else:
            self.log_fail(f"POST /login expected 200, got {status}")

        # GET /devices (authed)
        print("\n--- GET /devices ---")
        status, _, _ = self.req("/devices", with_cookie=True)
        if status is not None and status not in (401, 403):
            self.log_pass(f"GET /devices => {status}")
        else:
            self.log_fail(f"GET /devices expected auth success, got {status}")

        # POST /devices/createProvision
        print("\n--- POST /devices/createProvision ---")
        status, _, _ = self.req("/devices/createProvision?displayName=SmokeDevice", method="POST", with_cookie=True)
        if status is not None and status not in (401, 403):
            self.log_pass(f"POST /devices/createProvision => {status}")
        else:
            self.log_fail(f"POST /devices/createProvision expected auth success, got {status}")

        # POST /logout
        print("\n--- POST /logout ---")
        status, _, _ = self.req("/logout", method="POST", with_cookie=True)
        if status == 200:
            self.log_pass("POST /logout => 200")
        else:
            self.log_fail(f"POST /logout expected 200, got {status}")

        # GET /devices after logout (should fail)
        print("\n--- GET /devices (after logout) ---")
        status, _, _ = self.req("/devices", with_cookie=True)
        if status in (401, 403):
            self.log_pass(f"GET /devices after logout => {status} (expected denied)")
        else:
            self.log_fail(f"GET /devices after logout expected 401/403, got {status}")

        # POST /pw/reset
        print("\n--- POST /pw/reset ---")
        status, _, _ = self.req(f"/pw/reset?email={self.user_email}", method="POST")
        if status == 200:
            self.log_pass("POST /pw/reset => 200")
        else:
            self.log_fail(f"POST /pw/reset expected 200, got {status}")

        # run data endpoints only if we had a valid device / db scenario
        print("\n--- GET /data/capacitance (known to need existing data) ---")
        status, _, _ = self.req("/data/capacitance?deviceId=1&startTime=2024-01-01", with_cookie=True)
        # if not authenticated or internal error, we still note it
        if status is not None and status not in (401, 403):
            self.log_pass(f"GET /data/capacitance => {status}")
        else:
            self.log_fail(f"GET /data/capacitance expected authed, got {status}")

        print("\n--- GET /data/temperature (known to need existing data) ---")
        status, _, _ = self.req("/data/temperature?deviceId=1&startTime=2024-01-01", with_cookie=True)
        if status is not None and status not in (401, 403):
            self.log_pass(f"GET /data/temperature => {status}")
        else:
            self.log_fail(f"GET /data/temperature expected authed, got {status}")

        # summary
        print()
        total = self.passed + self.failed
        print(f"=== Results: {self.passed}/{total} passed, {self.failed} failed ===")
        return self.failed == 0


if __name__ == "__main__":
    args = parse_args()
    runner = SmokeTestRunner(args.url, verbose=args.verbose)
    ok = runner.run()
    sys.exit(0 if ok else 1)
