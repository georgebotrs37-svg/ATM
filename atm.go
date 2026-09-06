package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ============================================================
// Global constants
// ============================================================

const dataFile = "accounts.json"
const maxPinAttempts = 3
const dailyWithdrawLimit = 5000.0

// ============================================================
// Models: Transaction and Account
// ============================================================

// Transaction represents a single operation on an account (deposit / withdraw / transfer)
type Transaction struct {
	Type      string
	Amount    float64
	Balance   float64
	Timestamp string
}

// Account represents a customer's account in the ATM system
type Account struct {
	CardNumber      string
	PIN             string
	Owner           string
	Balance         float64
	Transactions    []Transaction
	WithdrawnToday  float64
	LastWithdrawDay string
}

func (a *Account) addTransaction(kind string, amount float64) {
	a.Transactions = append(a.Transactions, Transaction{
		Type:      kind,
		Amount:    amount,
		Balance:   a.Balance,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	})
}

// Deposit adds an amount to the account balance
func (a *Account) Deposit(amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	a.Balance += amount
	a.addTransaction("Deposit", amount)
	return nil
}

func (a *Account) resetDailyLimitIfNewDay() {
	today := time.Now().Format("2006-01-02")
	if a.LastWithdrawDay != today {
		a.LastWithdrawDay = today
		a.WithdrawnToday = 0
	}
}

// Withdraw takes an amount out of the account, respecting balance and the daily withdrawal limit
func (a *Account) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if amount > a.Balance {
		return errors.New("insufficient balance to complete this operation")
	}
	a.resetDailyLimitIfNewDay()
	if a.WithdrawnToday+amount > dailyWithdrawLimit {
		return fmt.Errorf("daily withdrawal limit exceeded (%.2f)", dailyWithdrawLimit)
	}
	a.Balance -= amount
	a.WithdrawnToday += amount
	a.addTransaction("Withdraw", amount)
	return nil
}

// ChangePIN changes the PIN after verifying the old one
func (a *Account) ChangePIN(oldPin, newPin string) error {
	if a.PIN != oldPin {
		return errors.New("current PIN is incorrect")
	}
	if len(newPin) != 4 {
		return errors.New("new PIN must be exactly 4 digits")
	}
	a.PIN = newPin
	return nil
}

// ============================================================
// Storage: save and load accounts from a JSON file
// ============================================================

// LoadAccounts loads accounts from a JSON file, or creates default accounts if the file doesn't exist
func LoadAccounts(path string) (map[string]*Account, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultAccounts(), nil
		}
		return nil, err
	}
	var accounts map[string]*Account
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

// SaveAccounts persists accounts to a JSON file
func SaveAccounts(path string, accounts map[string]*Account) error {
	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func defaultAccounts() map[string]*Account {
	return map[string]*Account{
		"1111": {CardNumber: "1111", PIN: "1234", Owner: "John Smith", Balance: 5000},
		"2222": {CardNumber: "2222", PIN: "4321", Owner: "Sarah Johnson", Balance: 12000},
	}
}

// ============================================================
// ATM: menus and user interaction
// ============================================================

// ATM represents a full running session of the machine
type ATM struct {
	accounts map[string]*Account
	current  *Account
	reader   *bufio.Reader
}

// NewATM creates a new ATM and loads account data
func NewATM() (*ATM, error) {
	accounts, err := LoadAccounts(dataFile)
	if err != nil {
		return nil, err
	}
	return &ATM{
		accounts: accounts,
		reader:   bufio.NewReader(os.Stdin),
	}, nil
}

func (m *ATM) readLine(prompt string) string {
	fmt.Print(prompt)
	text, _ := m.reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func (m *ATM) readAmount(prompt string) (float64, error) {
	s := m.readLine(prompt)
	amount, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errors.New("invalid value, please enter a number")
	}
	return amount, nil
}

// Login asks for card number and PIN, with a maximum of 3 attempts
func (m *ATM) Login() bool {
	for attempt := 1; attempt <= maxPinAttempts; attempt++ {
		card := m.readLine("Card number: ")
		acc, ok := m.accounts[card]
		if !ok {
			fmt.Println("✗ Card number not found")
			continue
		}
		pin := m.readLine("PIN: ")
		if acc.PIN == pin {
			m.current = acc
			fmt.Printf("\n✓ Welcome, %s\n", acc.Owner)
			return true
		}
		fmt.Printf("✗ Incorrect PIN (attempt %d of %d)\n", attempt, maxPinAttempts)
	}
	fmt.Println("Session locked: too many failed attempts.")
	return false
}

func (m *ATM) showMenu() {
	fmt.Println(`
============ ATM MENU ============
1. Check balance
2. Deposit money
3. Withdraw money
4. Transfer to another account
5. View account statement
6. Change PIN
7. Log out
===================================`)
}

// Run starts the ATM main loop after login
func (m *ATM) Run() {
	fmt.Println("=== Welcome to the ATM system ===")
	if !m.Login() {
		return
	}
	for {
		m.showMenu()
		choice := m.readLine("Choose an option: ")
		switch choice {
		case "1":
			fmt.Printf("Current balance: %.2f\n", m.current.Balance)
		case "2":
			m.deposit()
		case "3":
			m.withdraw()
		case "4":
			m.transfer()
		case "5":
			m.printStatement()
		case "6":
			m.changePin()
		case "7":
			fmt.Println("Logged out successfully. Thank you for using the ATM.")
			SaveAccounts(dataFile, m.accounts)
			return
		default:
			fmt.Println("✗ Invalid option, please try again.")
		}
	}
}

func (m *ATM) deposit() {
	amount, err := m.readAmount("Enter deposit amount: ")
	if err != nil {
		fmt.Println("✗ Error:", err)
		return
	}
	if err := m.current.Deposit(amount); err != nil {
		fmt.Println("✗ Error:", err)
		return
	}
	fmt.Printf("✓ Deposit successful. Current balance: %.2f\n", m.current.Balance)
	SaveAccounts(dataFile, m.accounts)
}

func (m *ATM) withdraw() {
	amount, err := m.readAmount("Enter withdrawal amount: ")
	if err != nil {
		fmt.Println("✗ Error:", err)
		return
	}
	if err := m.current.Withdraw(amount); err != nil {
		fmt.Println("✗ Error:", err)
		return
	}
	fmt.Printf("✓ Withdrawal successful. Current balance: %.2f\n", m.current.Balance)
	SaveAccounts(dataFile, m.accounts)
}

func (m *ATM) transfer() {
	destCard := m.readLine("Enter recipient's card number: ")
	dest, ok := m.accounts[destCard]
	if !ok {
		fmt.Println("✗ Card number not found")
		return
	}
	if dest.CardNumber == m.current.CardNumber {
		fmt.Println("✗ Cannot transfer to the same account")
		return
	}
	amount, err := m.readAmount("Enter transfer amount: ")
	if err != nil {
		fmt.Println("✗ Error:", err)
		return
	}
	if err := m.current.Withdraw(amount); err != nil {
		fmt.Println("✗ Error:", err)
		return
	}
	if err := dest.Deposit(amount); err != nil {
		// Extremely unlikely since amount was already validated, but refund if it ever happens
		_ = m.current.Deposit(amount)
		fmt.Println("✗ Error:", err)
		return
	}
	fmt.Printf("✓ Transferred %.2f to %s successfully.\n", amount, dest.Owner)
	SaveAccounts(dataFile, m.accounts)
}

func (m *ATM) printStatement() {
	fmt.Println("\n--- Account Statement ---")
	if len(m.current.Transactions) == 0 {
		fmt.Println("No transactions yet.")
		return
	}
	for _, t := range m.current.Transactions {
		fmt.Printf("[%s] %-9s Amount: %-10.2f Balance after: %.2f\n",
			t.Timestamp, t.Type, t.Amount, t.Balance)
	}
}

func (m *ATM) changePin() {
	oldPin := m.readLine("Enter current PIN: ")
	newPin := m.readLine("Enter new PIN (4 digits): ")
	if err := m.current.ChangePIN(oldPin, newPin); err != nil {
		fmt.Println("✗ Error:", err)
		return
	}
	fmt.Println("✓ PIN changed successfully.")
	SaveAccounts(dataFile, m.accounts)
}

// ============================================================
// Program entry point
// ============================================================

func main() {
	atm, err := NewATM()
	if err != nil {
		fmt.Println("System startup error:", err)
		return
	}
	atm.Run()
}
