#!/usr/bin/env python3
"""
Creates Stripe test prices for each product and sets stripe_price_id via admin API.

Required environment variables (export before running):
  STRIPE_SECRET_KEY   — Stripe test secret key (sk_test_...)
  ADMIN_SECRET        — X-Admin-Secret header value  (default: gour-admin-dev-secret)
  API_BASE            — backend URL                  (default: http://localhost:8080/api/v1)

Run:
  STRIPE_SECRET_KEY=sk_test_... python3 scripts/setup_stripe_prices.py
"""

import json
import os
import sys
import urllib.request
import urllib.parse
import urllib.error
import base64

STRIPE_SK    = os.getenv("STRIPE_SECRET_KEY", "")
ADMIN_SECRET = os.getenv("ADMIN_SECRET", "gour-admin-dev-secret")
API_BASE     = os.getenv("API_BASE", "http://localhost:8080/api/v1")

if not STRIPE_SK:
    print("ERROR: STRIPE_SECRET_KEY environment variable not set.", file=sys.stderr)
    sys.exit(1)

PRODUCTS = [
    {"name": "cold_email", "display": "Cold Email",  "amount": 2900},
    {"name": "linkedin",   "display": "LinkedIn",    "amount": 1900},
    {"name": "warmup",     "display": "Email Warmup","amount": 990},
    {"name": "sendflow",   "display": "SendFlow",    "amount": 2900},
]


def stripe_post(path, params):
    auth = base64.b64encode(f"{STRIPE_SK}:".encode()).decode()
    data = urllib.parse.urlencode(params).encode()
    req = urllib.request.Request(
        f"https://api.stripe.com{path}",
        data=data,
        headers={"Authorization": f"Basic {auth}",
                 "Content-Type": "application/x-www-form-urlencoded"},
    )
    with urllib.request.urlopen(req, timeout=20) as r:
        return json.loads(r.read())


def admin_patch(product_id, stripe_price_id):
    data = json.dumps({"stripe_price_id": stripe_price_id}).encode()
    req = urllib.request.Request(
        f"{API_BASE}/admin/products/{product_id}",
        data=data,
        method="PATCH",
        headers={"X-Admin-Secret": ADMIN_SECRET,
                 "Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=10) as r:
        return json.loads(r.read())


def admin_get_products():
    req = urllib.request.Request(
        f"{API_BASE}/admin/products",
        headers={"X-Admin-Secret": ADMIN_SECRET},
    )
    with urllib.request.urlopen(req, timeout=10) as r:
        return json.loads(r.read())["products"]


def main():
    print("Fetching products from admin API...")
    products_db = {p["name"]: p["id"] for p in admin_get_products()}
    print(f"  Found: {list(products_db.keys())}")

    for p in PRODUCTS:
        name = p["name"]
        if name not in products_db:
            print(f"  ⚠  Product '{name}' not found in DB, skipping")
            continue

        print(f"\nCreating Stripe price for '{name}' (${p['amount']/100:.2f}/mo)...")
        price = stripe_post("/v1/prices", {
            "currency": "usd",
            "unit_amount": p["amount"],
            "recurring[interval]": "month",
            "product_data[name]": p["display"],
        })
        price_id = price["id"]
        print(f"  ✅ Stripe price created: {price_id}")

        print(f"  Patching product {products_db[name]}...")
        result = admin_patch(products_db[name], price_id)
        stored = result.get("product", {}).get("stripe_price_id") or result.get("stripe_price_id")
        print(f"  ✅ stripe_price_id stored: {stored}")

    print("\n🎉 All products have Stripe prices. Ready to test checkout!")


if __name__ == "__main__":
    main()
