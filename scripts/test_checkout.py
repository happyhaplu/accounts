#!/usr/bin/env python3
"""
E2E test: login → list workspaces → create checkout session → verify response.

Required environment variables (or set defaults below):
  ADMIN_SECRET   — X-Admin-Secret value  (default: outcraftly-admin-dev-secret)
  TEST_EMAIL     — test user email        (default: test@example.com)
  TEST_PASSWORD  — test user password     (default: Test@1234)
  API_BASE       — backend URL            (default: http://localhost:8080/api/v1)

Run:
  python3 scripts/test_checkout.py
"""

import json
import os
import urllib.request
import urllib.parse
import urllib.error

API = os.getenv("API_BASE", "http://localhost:8080/api/v1")

EMAIL    = os.getenv("TEST_EMAIL",    "test@example.com")
PASSWORD = os.getenv("TEST_PASSWORD", "Test@1234")

ADMIN_SECRET = os.getenv("ADMIN_SECRET", "outcraftly-admin-dev-secret")


def post(path, data, token=None):
    body = json.dumps(data).encode()
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    req = urllib.request.Request(f"{API}{path}", data=body, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            return r.status, json.loads(r.read())
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read())


def get(path, token=None):
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    req = urllib.request.Request(f"{API}{path}", headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            return r.status, json.loads(r.read())
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read())


def admin_get(path):
    req = urllib.request.Request(f"{API}{path}",
                                  headers={"X-Admin-Secret": ADMIN_SECRET})
    with urllib.request.urlopen(req, timeout=10) as r:
        return json.loads(r.read())


def main():
    # 1. Register a test user (ignore error if already exists)
    print("1. Registering test user...")
    status, body = post("/auth/register", {"email": EMAIL, "password": PASSWORD, "name": "Test User"})
    print(f"   → {status}: {body.get('message') or body.get('error') or body}")

    # 2. Login
    print("2. Logging in...")
    status, body = post("/auth/login", {"email": EMAIL, "password": PASSWORD})
    if status != 200:
        print(f"   ✗ Login failed: {body}")
        return
    token = body["token"]
    print(f"   ✅ token: {token[:30]}...")

    # 3. List workspaces
    print("3. Listing workspaces...")
    status, body = get("/workspaces", token)
    if status != 200:
        print(f"   ✗ {status}: {body}")
        # Try to create one
        print("   Creating workspace...")
        status, body = post("/workspaces", {"name": "Test Workspace"}, token)
        if status != 201:
            print(f"   ✗ Create workspace failed: {body}")
            return
        ws_id = body["workspace"]["id"]
    else:
        workspaces = body.get("workspaces", [])
        if not workspaces:
            print("   No workspaces, creating one...")
            s2, b2 = post("/workspaces", {"name": "Test Workspace"}, token)
            ws_id = b2["workspace"]["id"]
        else:
            ws_id = workspaces[0]["id"]
    print(f"   ✅ workspace_id: {ws_id}")

    # 4. Get products
    print("4. Getting products with stripe_price_id...")
    products = admin_get("/admin/products")["products"]
    product = next((p for p in products if p.get("stripe_price_id")), None)
    if not product:
        print("   ✗ No product with stripe_price_id found! Run setup_stripe_prices.py first.")
        return
    print(f"   ✅ using product: {product['name']} (price: {product['stripe_price_id']})")

    # 5. Create checkout session
    print("5. Creating Stripe checkout session...")
    status, body = post(
        f"/workspaces/{ws_id}/billing/checkout",
        {
            "price_id":    product["stripe_price_id"],
            "product_id":  product["id"],
            "plan_name":   product["name"],
            "success_url": "http://localhost:5173/billing?success=true",
            "cancel_url":  "http://localhost:5173/billing?canceled=true",
        },
        token,
    )
    if status != 200:
        print(f"   ✗ {status}: {body}")
        return
    url = body.get("url", "")
    print(f"   ✅ Checkout URL: {url[:80]}...")
    print()
    print("🎉 E2E checkout session test PASSED!")
    print(f"   Open this URL in browser to complete payment: {url}")


if __name__ == "__main__":
    main()
