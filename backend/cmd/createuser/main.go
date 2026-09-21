// Command createuser creates a CampusWatch operator account.
//
// CampusWatch has no public registration — access is restricted to authorised
// operators, and every dashboard route sits behind RequireRole("admin",
// "manager") in internal/server. That leaves a bootstrapping problem: the
// dashboard is unusable until an account exists, but nothing in the running
// system can create one. This command is the way in.
//
// Run it from the repository root:
//
//	go run ./backend/cmd/createuser \
//	    --email admin@adewah.edu \
//	    --password 'a-strong-password' \
//	    --institution 'Adewah University'
//
// The institution is created on first use and reused afterwards, so running
// this again with the same --institution adds another account rather than
// failing on the institutions.slug UNIQUE constraint.
//
// Pass the password through CAMPUSWATCH_PASSWORD rather than --password when
// the machine is shared: a flag is visible in the shell history and in `ps`.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/config"
	"CampusWatch/backend/internal/database"
	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/repository"
	"CampusWatch/backend/internal/service"
)

// validRoles mirrors the users_role_check constraint in
// migrations/007_users.sql. Validating here turns a database constraint
// violation into a message that names the acceptable values.
var validRoles = []string{"admin", "manager", "operator", "viewer"}

func main() {
	email := flag.String("email", "", "email address of the new account (required)")
	passwordFlag := flag.String("password", "", "password for the account; falls back to $CAMPUSWATCH_PASSWORD (required)")
	firstName := flag.String("first-name", "CampusWatch", "given name of the account holder")
	lastName := flag.String("last-name", "Administrator", "family name of the account holder")
	institutionName := flag.String("institution", "", "institution to attach the account to, created if absent (required)")
	role := flag.String("role", "admin", "role to grant: "+strings.Join(validRoles, ", "))
	flag.Parse()

	if err := run(*email, *passwordFlag, *firstName, *lastName, *institutionName, *role); err != nil {
		log.Fatal(err)
	}
}

// run performs the work, returning an error instead of exiting so that every
// failure path is visible in one place.
func run(email, passwordFlag, firstName, lastName, institutionName, role string) error {
	email = strings.TrimSpace(email)
	institutionName = strings.TrimSpace(institutionName)
	role = strings.TrimSpace(role)

	// A flag takes precedence, but the environment is the safer way to supply
	// this and is the documented alternative.
	passwordValue := passwordFlag
	passwordFromFlag := passwordValue != ""
	if !passwordFromFlag {
		passwordValue = os.Getenv("CAMPUSWATCH_PASSWORD")
	}

	switch {
	case email == "":
		return errors.New("--email is required")
	case passwordValue == "":
		return errors.New("a password is required: pass --password or set CAMPUSWATCH_PASSWORD")
	case institutionName == "":
		return errors.New("--institution is required")
	case !isValidRole(role):
		return fmt.Errorf("invalid --role %q; expected one of %s", role, strings.Join(validRoles, ", "))
	}

	// Only DATABASE_URL is needed here, so PORT is deliberately not required.
	databaseURL, err := config.LoadDatabaseURL()
	if err != nil {
		return err
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	// Apply migrations here rather than relying on the server having been
	// started first. Against a fresh database the users table does not exist
	// yet, and requiring an operator to start a server they do not need just to
	// create an account is a trap. ApplyMigrations records what it has already
	// run, so doing this again on every start costs one query.
	migrationDir, err := database.ResolveMigrationsDir()
	if err != nil {
		return err
	}
	if err := database.ApplyMigrations(ctx, db, migrationDir); err != nil {
		return err
	}

	institutionService := service.NewInstitutionService(repository.NewInstitutionRepository(db))
	userRepo := repository.NewUserRepository(db)

	institution, institutionCreated, err := ensureInstitution(ctx, institutionService, institutionName)
	if err != nil {
		return err
	}

	// Scoped to the institution: the same address may legitimately hold an
	// account at two institutions, and users_email_unique is (institution_id,
	// email) to allow exactly that.
	existing, err := userRepo.FindByEmailInInstitution(ctx, institution.ID, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf(
			"user %s already exists in %s (id %s); nothing was changed",
			email, institution.Name, existing.ID,
		)
	}

	hash, err := password.HashPassword(passwordValue)
	if err != nil {
		return err
	}

	created, err := userRepo.Create(ctx, model.User{
		InstitutionID: institution.ID,
		Email:         email,
		PasswordHash:  hash,
		FirstName:     firstName,
		LastName:      lastName,
		Role:          role,
		IsActive:      true,
	})
	if err != nil {
		return err
	}

	institutionState := "reused"
	if institutionCreated {
		institutionState = "created"
	}

	fmt.Printf("institution %s (%s) - %s\n", institution.Name, institution.Slug, institutionState)
	fmt.Printf("user        %s (%s) - created\n", created.Email, created.Role)
	fmt.Printf("user id     %s\n", created.ID)
	fmt.Println()
	fmt.Println("Sign in at http://localhost:8000/ with the backend and dashboard running.")

	if passwordFromFlag {
		fmt.Println()
		fmt.Println("Note: --password is recorded in your shell history. On a shared machine,")
		fmt.Println("prefer supplying it through CAMPUSWATCH_PASSWORD.")
	}

	return nil
}

// ensureInstitution returns the institution for name, creating it if absent.
//
// It reports whether it had to create one, so the caller can say which happened
// rather than leaving the operator to guess whether a row was added.
func ensureInstitution(ctx context.Context, svc *service.InstitutionService, name string) (*model.Institution, bool, error) {
	existing, err := svc.GetBySlug(ctx, name)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	// An empty slug and status let the service derive the slug from the name and
	// default the status, so the rules live in one place.
	created, err := svc.Create(ctx, name, "", "")
	if err != nil {
		return nil, false, err
	}

	return created, true, nil
}

func isValidRole(role string) bool {
	for _, candidate := range validRoles {
		if role == candidate {
			return true
		}
	}
	return false
}
