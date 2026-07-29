# SalesTrack API

Backend for the firecracker sales app: a mobile app for sales agents (create
bills, send to WhatsApp) and a web admin portal (approve bills, manage
products/staff, GST reports). Go + Gin + PostgreSQL.

## ⚠️ Before you run this

This code was written without access to a working Go toolchain (the sandbox
it was generated in ran out of disk space, so `go build` / `go mod tidy`
could never be executed here). That means:

1. **Run `go mod tidy` first.** `go.mod` lists the direct dependencies
   (gin, jwt, godotenv, razorpay-go, excelize, bcrypt) but `go.sum` only has
   the entries that existed before this change. `go mod tidy` will fetch
   everything and fill in `go.sum` correctly.
2. **Then `go build ./...`** to confirm everything compiles, and fix up
   anything that doesn't match your exact dependency versions (in
   particular, double-check `razorpay-go`'s `Order.Create` signature against
   [its README](https://github.com/razorpay/razorpay-go) — that's the one
   third-party call I couldn't cross-check without network access).
3. Run the migration (`migrations/001_init.sql`) against a real Postgres
   database before starting the server.

## Setup

```bash
cp .env.example .env        # fill in DATABASE_URL, JWT_SECRET, Razorpay keys
createdb salestrack
psql "$DATABASE_URL" -f migrations/001_init.sql
go mod tidy
go run ./cmd/api
```

Bootstrap your shop's first admin login (there's no UI for this — it's a
one-time setup call):

```bash
curl -X POST localhost:8080/api/admin/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Sunil Milwani","mobile_number":"9822012345","password":"changeme123","shop_name":"Shama Fireworks","bill_prefix":"SF"}'
```

Then log in as that admin, add products (Product Management), and add sales
staff (User Management) — creating a staff member returns a 4-digit
`login_code` **once**; that's what the agent enters on the mobile login
screen alongside their mobile number.

## How the pieces map to the screens you shared

| Screen | Endpoint(s) |
|---|---|
| Sales staff login (mobile + PIN) | `POST /api/sales/auth/login` |
| Home (today's bills/sales, recent bills) | `GET /api/sales/home` |
| New Bill → product picker | `GET /api/sales/products?search=&category=` |
| New Bill → submit / Confirmation | `POST /api/sales/bills` |
| Send to WhatsApp | `POST /api/sales/bills/:id/whatsapp` |
| Bill History (Today/Week/All) | `GET /api/sales/bills?range=today\|week\|all` |
| Admin login | `POST /api/admin/auth/login` |
| Dashboard "Today's overview" | `GET /api/admin/dashboard` |
| Product Management | `GET/POST/PUT/DELETE /api/admin/products` + `PATCH /:id/status` |
| User Management | `GET/POST/PUT/DELETE /api/admin/staff` + `PATCH /:id/status`, `POST /:id/reset-code` |
| Cash Counter — pending approvals | `GET /api/admin/bills/pending` |
| Cash Counter — approve (Cash / UPI) | `POST /api/admin/bills/:id/approve {"payment_mode":"cash"\|"upi"}` |
| Report Management — Bill-wise | `GET /api/admin/reports/bills?from=&to=&staff_id=` |
| Report Management — Product-wise | `GET /api/admin/reports/products?from=&to=&staff_id=` |
| Report Management — Download Excel | `GET /api/admin/reports/download?type=bill\|product&from=&to=&staff_id=` |

All `/api/admin/*` and `/api/sales/*` routes (except `/auth/*`) require
`Authorization: Bearer <token>` from the matching login call.

## Design decisions worth knowing about

- **"OTP" login is actually the admin-issued 4-digit PIN shown in User
  Management.** Your screens show a PIN keypad, not an SMS OTP flow, and
  User Management auto-generates a code per staff member — so that's what's
  wired up. It's mocked in the sense that there's no SMS provider involved;
  swapping in real SMS OTP later just means adding a send/verify step in
  front of `auth.Service.LoginSalesStaff`, the rest doesn't change.
- **UPI payments go through Razorpay.** Tapping "UPI" on Cash Counter opens
  a Razorpay order (`payment.Service.CreateOrderForBill`) and returns
  `order_id`/`key_id`/`amount` for the admin webapp to render Razorpay
  Checkout to the customer. The bill only flips to `approved` once payment
  is confirmed — either immediately via `POST /bills/:id/verify-payment`
  (Checkout.js's success callback) or authoritatively via the
  `/api/payments/razorpay/webhook` endpoint (add that URL + a webhook
  secret in the Razorpay dashboard). Tapping "Cash" approves instantly, no
  gateway involved.
- **"Send to WhatsApp" generates a `wa.me` deep link**, not a WhatsApp
  Business API call — no WhatsApp API credentials needed. The agent's phone
  opens WhatsApp with the bill summary prefilled and sends it themselves.
  Swap `internal/whatsapp/link.go` for the WhatsApp Cloud API later if you
  want the server to send it automatically.
- **GST**: every product stores a taxable value + GST%; CGST/SGST are always
  the GST% split evenly (matches "CGST @ 9%, SGST @ 9%" for an 18% slab on
  your GST Snapshot screen). Bill numbers look like `SF/26-27/00483` —
  prefix, Indian financial year, then a per-shop running sequence.

## Project layout

```
cmd/api/main.go       entrypoint, wires everything, starts Gin
config/                env config + Postgres connection pool
migrations/001_init.sql  schema
internal/
  auth/                 admin + sales-staff login, JWT
  staff/                admin's "User Management" (sales staff CRUD)
  product/               "Product Management" + sales-side product listing
  billing/               core Bill/BillItem model — bill creation, approval, stats
  payment/                Razorpay order creation + webhook
  whatsapp/               wa.me link builder
  admin/                  dashboard aggregation + "Cash Counter" routes
  sales/                  mobile app routes (home, bills, whatsapp)
  report/                 Bill-wise / Product-wise reports + Excel export
  middleware/             JWT auth guard, CORS
pkg/
  response/               standard {success,message,data} JSON envelope
  utils/                  GST math, financial-year formatting
```
