package tests

import (
	"context"
	"log"
	"math/rand/v2"
	"os"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/joho/godotenv"
	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/permissions"
	"github.com/prior-it/apollo/postgres"
)

var Faker = gofakeit.New(rand.Uint64())

func DB() *postgres.ApolloDB {
	ctx := context.Background()
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Could not load the .env file: %v", err)
	}
	url := os.Getenv("DATABASE_URL")
	db, err := postgres.NewDB(ctx, url)
	if err != nil {
		panic(
			"To test database functionality, set the DATABASE_URL env variable to a valid database",
		)
	}

	err = db.MigrateApollo(ctx)
	if err != nil {
		log.Panicf("Cannot migrate apollo db: %v", err)
	}

	return db
}

func DeleteAllUsers(service core.UserService) {
	ctx := context.Background()
	users, err := service.ListUsers(ctx)
	Check(err)
	for _, user := range users {
		Check(service.DeleteUser(ctx, user.ID))
	}
}

func DeleteAllOrganisations(service core.OrganisationService) {
	ctx := context.Background()
	organisations, err := service.ListOrganisations(ctx)
	Check(err)
	for _, organisation := range organisations {
		Check(service.DeleteOrganisation(ctx, organisation.ID))
	}
}

func DeleteAllAddresses(service core.AddressService) {
	ctx := context.Background()
	addresss, err := service.ListAddresses(ctx)
	Check(err)
	for _, address := range addresss {
		Check(service.DeleteAddress(ctx, address.ID))
	}
}

func DeleteAllPermissions(service permissions.Service) {
	ctx := context.Background()
	groups, err := service.ListPermissionGroups(ctx)
	Check(err)
	for _, group := range groups {
		Check(service.DeletePermissionGroup(ctx, group.ID))
	}
}

func CreateRegularUser(service core.UserService) *core.User {
	email, err := core.NewEmailAddress(Faker.Email())
	if err != nil {
		log.Fatal(err)
	}
	user, err := service.CreateUser(context.Background(), Faker.Name(), email)
	if err != nil {
		log.Fatalf("cannot create regular user: %v", err)
	}
	return user
}

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
