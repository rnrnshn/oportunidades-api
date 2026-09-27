package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/mail"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/rnrnshn/oportunidades-api/pkg/db"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

const minPasswordLength = 12

type input struct {
	Email    string
	Name     string
	Password string
}

func main() {
	in, err := parseInput(os.Args[1:], os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create-admin:", err)
		os.Exit(2)
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "create-admin: DATABASE_URL is required")
		os.Exit(2)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create-admin: connect:", err)
		os.Exit(1)
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create-admin: hash password:", err)
		os.Exit(1)
	}

	user, err := queries.New(pool).UpsertAdminUser(ctx, queries.UpsertAdminUserParams{
		Email:        in.Email,
		PasswordHash: string(hash),
		Name:         in.Name,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "create-admin: save user:", err)
		os.Exit(1)
	}

	fmt.Printf("admin ready: %s (%s)\n", user.Email, user.Name)
}

// The password is read from stdin so it never appears in shell history or process listings.
func parseInput(args []string, stdin io.Reader) (input, error) {
	flags := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	email := flags.String("email", "", "admin email")
	name := flags.String("name", "", "admin display name")
	if err := flags.Parse(args); err != nil {
		return input{}, fmt.Errorf("%w (usage: create-admin --email EMAIL --name NAME, password on stdin)", err)
	}

	in := input{
		Email: strings.ToLower(strings.TrimSpace(*email)),
		Name:  strings.TrimSpace(*name),
	}
	if _, err := mail.ParseAddress(in.Email); err != nil || in.Email == "" {
		return input{}, errors.New("--email must be a valid email address")
	}
	if in.Name == "" {
		return input{}, errors.New("--name is required")
	}

	line, err := bufio.NewReader(stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return input{}, fmt.Errorf("read password: %w", err)
	}
	in.Password = strings.TrimRight(line, "\r\n")
	if len(in.Password) < minPasswordLength {
		return input{}, fmt.Errorf("password on stdin must be at least %d characters", minPasswordLength)
	}

	return in, nil
}
