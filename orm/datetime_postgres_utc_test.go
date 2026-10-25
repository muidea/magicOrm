//go:build !mysql

package orm

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/utils"
)

type utcStorageSample struct {
	ID       int64      `orm:"id key auto" view:"detail,lite"`
	Stamp    time.Time  `orm:"stamp" view:"detail,lite"`
	Optional *time.Time `orm:"optional" view:"detail,lite"`
}

func utcStorageFixture(t *testing.T, kind string, value *utcStorageSample) (provider.Provider, models.Model) {
	t.Helper()
	p := provider.NewLocalProvider("utc-storage", nil)
	var definition any = &utcStorageSample{}
	var input any = value
	if kind == "remote" {
		p = provider.NewRemoteProvider("utc-storage", nil)
		object, err := helper.GetObject(definition)
		if err != nil {
			t.Fatal(err)
		}
		definition = object
		objectValue, err := helper.GetObjectValue(value)
		if err != nil {
			t.Fatal(err)
		}
		input = objectValue
	}
	if _, err := p.RegisterModel(definition); err != nil {
		t.Fatal(err)
	}
	model, err := p.GetEntityModel(input, true)
	if err != nil {
		t.Fatal(err)
	}
	return p, model
}

// Only random-prefix tables in an explicitly supplied disposable PostgreSQL.
// All DDL, inserts, updates, queries and cleanup use the public ORM contract.
func TestDateTimeUTCIsolatedPostgres(t *testing.T) {
	if os.Getenv("MAGICORM_DATETIME_TEST_ALLOW") != "isolated" {
		t.Skip("set isolated datetime PostgreSQL configuration")
	}
	server, name, user, password := os.Getenv("MAGICORM_DATETIME_TEST_SERVER"), os.Getenv("MAGICORM_DATETIME_TEST_DATABASE"), os.Getenv("MAGICORM_DATETIME_TEST_USER"), os.Getenv("MAGICORM_DATETIME_TEST_PASSWORD")
	if server == "" || !strings.HasPrefix(name, "magicorm_test_") || user == "" || password == "" {
		t.Fatal("isolated datetime configuration incomplete")
	}
	schema := os.Getenv("MAGICORM_DATETIME_TEST_SCHEMA")
	if schema == "" || schema == "public" {
		t.Fatal("dedicated non-public schema required")
	}
	cfg := NewConfig(server, name, schema, user, password)
	for _, writerKind := range []string{"local", "remote"} {
		t.Run(writerKind, func(t *testing.T) {
			prefix := "utc_" + strings.ToLower(rand.Text())
			p, template := utcStorageFixture(t, writerKind, &utcStorageSample{})
			writer, err := NewOrm(p, cfg, prefix)
			if err != nil {
				t.Fatal("isolated writer unavailable", err.Code)
			}
			defer writer.Release()
			if err := writer.Create(template); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := writer.Drop(template); err != nil {
					t.Error(err)
					return
				}
				report, err := writer.InspectSchema(template, false)
				if err != nil || report == nil || !report.Matches {
					t.Error("owned datetime table cleanup unconfirmed", err)
				}
			}()
			for _, text := range []string{"2026-10-21T00:30:45.123+08:00", "2026-10-20T11:30:45.123-05:00", "2026-10-20T16:30:45.123Z"} {
				instant, parseErr := time.Parse(time.RFC3339Nano, text)
				if parseErr != nil {
					t.Fatal(parseErr)
				}
				_, input := utcStorageFixture(t, writerKind, &utcStorageSample{Stamp: instant})
				if writerKind == "remote" {
					if err := input.SetFieldValue("stamp", text); err != nil {
						t.Fatal(err)
					}
				}
				created, insertErr := writer.Insert(input)
				if insertErr != nil {
					t.Fatal("datetime insert failed", insertErr)
				}
				id, idErr := strconv.ParseInt(fmt.Sprint(created.GetPrimaryField().GetValue().Get()), 10, 64)
				if idErr != nil || id <= 0 {
					t.Fatal("invalid allocated ID")
				}

				for _, operation := range []string{"insert", "update"} {
					expectedInstant := instant
					expectedClock := "2026-10-20 16:30:45.123"
					if operation == "update" {
						expectedInstant = instant.Add(24 * time.Hour)
						expectedClock = "2026-10-21 16:30:45.123"
						_, update := utcStorageFixture(t, writerKind, &utcStorageSample{ID: id, Stamp: expectedInstant})
						if writerKind == "remote" {
							if err := update.SetFieldValue("stamp", expectedInstant.Format(time.RFC3339Nano)); err != nil {
								t.Fatal(err)
							}
						}
						updated, err := writer.Update(update)
						if err != nil || updated == nil || fmt.Sprint(updated.GetPrimaryField().GetValue().Get()) != strconv.FormatInt(id, 10) {
							t.Fatal("datetime update changed identity or failed", err)
						}
					}
					for _, readerKind := range []string{"local", "remote"} {
						readProvider, query := utcStorageFixture(t, readerKind, &utcStorageSample{ID: id})
						reader, err := NewOrm(readProvider, cfg, prefix)
						if err != nil {
							t.Fatal("independent reader unavailable", err.Code)
						}
						result, err := reader.Query(query)
						reader.Release()
						if err != nil {
							t.Fatal("independent datetime query failed", err)
						}
						raw := result.GetField("stamp").GetValue().Get()
						decoded, err := utils.ConvertRawToDateTime(raw)
						if err != nil || !decoded.Equal(expectedInstant) || decoded.Location() != time.UTC {
							t.Fatal("persisted datetime is not the same UTC instant", raw, err)
						}
						if readerKind == "remote" && fmt.Sprint(raw) != expectedClock {
							t.Fatal("raw stored clock differs from UTC", raw)
						}
						if value := result.GetField("optional").GetValue(); value != nil && !value.IsZero() {
							t.Fatal("nullable datetime changed")
						}
					}
				}
			}
			t.Log("three offset inputs, inserts and updates, allocated IDs, independent local/remote reads and exact table cleanup verified")
		})
	}
}
