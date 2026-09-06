# ATM System (Go)

A simple command-line ATM simulator written in Go, with no external dependencies. Accounts, balances, and transaction history are persisted to a local JSON file so nothing is lost between runs.

## Features

- **Login** with card number and PIN (locks out after 3 failed attempts)
- **Check balance**
- **Deposit** money
- **Withdraw** money, with a daily withdrawal limit of 5000
- **Transfer** funds to another account
- **Account statement** showing every past transaction with timestamps
- **Change PIN**
- **Persistent storage** — all data is saved automatically to `accounts.json`

## Requirements

- [Go](https://go.dev/dl/) 1.18 or later

## Getting Started

1. Make sure `atm.go` is in its own project folder.
2. Open a terminal in that folder and initialize the Go module (only needed once):

   ```bash
   go mod init atm
   ```

3. Run the program:

   ```bash
   go run atm.go
   ```

   Or build a standalone executable:

   ```bash
   go build -o atm.exe atm.go   # Windows
   go build -o atm atm.go       # macOS/Linux
   ./atm
   ```

## Default Test Accounts

On first run, `accounts.json` doesn't exist yet, so the program creates two sample accounts automatically:

| Card Number | PIN  | Owner         | Balance |
|-------------|------|---------------|---------|
| 1111        | 1234 | John Smith    | 5000    |
| 2222        | 4321 | Sarah Johnson | 12000   |

## Project Structure

```
atm.go          # entire application: models, storage, ATM logic, entry point
accounts.json   # auto-generated after the first run; stores all account data
```

## How It Works

- **Account** holds the card number, PIN, owner name, balance, transaction history, and daily withdrawal tracking.
- **Storage** (`LoadAccounts` / `SaveAccounts`) reads and writes all accounts as JSON.
- **ATM** drives the login flow and the interactive menu loop, calling into `Account` methods for each operation and saving to disk after every change.

## Notes

- To reset all data, delete `accounts.json` and restart the program — the defaults above will be recreated.
- The daily withdrawal limit resets automatically at the start of a new calendar day.
