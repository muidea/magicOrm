package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/validation"
	verrors "github.com/muidea/magicOrm/validation/errors"
)

type Device struct {
	ID   int64  `orm:"id key auto" view:"detail,lite"`
	Name string `orm:"name" constraint:"req,min=2" view:"detail,lite"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// Return errors before main exits so all acquired resources are released.
func run() error {
	server, name, schema := os.Getenv("ORM_SERVER"), os.Getenv("ORM_DATABASE"), os.Getenv("ORM_SCHEMA")
	user, password := os.Getenv("ORM_USER"), os.Getenv("ORM_PASSWORD")
	if server == "" || name == "" || schema == "" || user == "" || password == "" {
		return fmt.Errorf("explicit ORM database configuration required")
	}
	// Schema provisioning is explicit and belongs to the deployment owner.
	if err := orm.EnsureSchema(server, name, schema, user, password); err != nil {
		return err
	}
	orm.Initialize()
	defer orm.Uninitialized()
	if err := orm.AddDatabase(server, name, schema, user, password, 4, "quickstart"); err != nil {
		return err
	}
	defer orm.DelDatabase("quickstart")
	p := provider.NewLocalProvider("quickstart", nil)
	model, err := p.RegisterModel(&Device{})
	if err != nil {
		return err
	}
	h, err := orm.GetOrm(context.Background(), p, "example")
	if err != nil {
		return err
	}
	defer h.Release()
	if err := h.Create(model); err != nil {
		return err
	}
	value, err := p.GetEntityModel(&Device{Name: "pump"}, true)
	if err != nil {
		return err
	}
	manager := validation.NewValidationManager(validation.DefaultConfig())
	ctx := validation.NewContext(verrors.ScenarioInsert, validation.OperationCreate, nil, "postgresql")
	if err := manager.ValidateModel(value, ctx); err != nil {
		return err
	}
	created, err := h.Insert(value)
	if err != nil {
		return err
	}
	log.Printf("created device %v", created.GetPrimaryField().GetValue().Get())
	return nil
}
