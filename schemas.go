package main

import (
	"context"
	"errors"
	"hepatitis-antiviral/cli"
	"hepatitis-antiviral/migrations"
	"hepatitis-antiviral/sources/mongo"
	"os"

	"github.com/infinitybotlist/eureka/crypto"
)

// Schemas here
//
// Either use schema struct tag (or bson + mark struct tag for special type overrides)

// Tool below

var source mongo.MongoSource

type UUID = string

type Server struct {
	ServerID string `src:"serverId" dest:"server_id" unique:"true"`
	Invite   string `src:"serverInvite" dest:"invite"`
}

type User struct {
	UserID string `src:"id" dest:"user_id" unique:"true"`
	Token  string `src:"token" dest:"token"`
	Banned bool   `src:"banned" dest:"banned" default:"false"`
	Staff  bool   `src:"staff" dest:"staff" default:"false"`
	Bio    string `src:"bio" dest:"bio" default:"'No bio set!'"`
}

var userTransforms = map[string]cli.TransformFunc{
	"Token": func(a cli.TransformRow) any {
		return crypto.RandString(255)
	},
}

type Template struct {
	Title       string `src:"title" dest:"title"`
	Image       string `src:"image" dest:"image"`
	Description string `src:"description" dest:"description" default:"'A simple widget template'"`
}

type Widget struct {
	ServerID string `src:"serverId" dest:"server_id" fk:"servers,server_id"`
	UserID   string `src:"userId" dest:"user_id" fk:"users,user_id"`
	Theme    string `src:"theme" dest:"theme" default:"'dark'"`
	Image    string `src:"image" dest:"image" default:"'/landing.svg'"`
	Domain   string `src:"domain" dest:"domain"`
	Token    string `src:"token" dest:"token"`
}

var widgetTransforms = map[string]cli.TransformFunc{
	"Token": func(a cli.TransformRow) any {
		return crypto.RandString(64)
	},
}

func main() {
	// Place all schemas to be used in the tool here

	cli.Main(cli.App{
		SchemaOpts: cli.SchemaOpts{
			TableName: "diswidgets",
		},
		// Required
		LoadSource: func(name string) (cli.Source, error) {
			switch name {
			case "mongo":
				source = mongo.MongoSource{
					ConnectionURL:  os.Getenv("MONGO"),
					DatabaseName:   "diswidgets",
					IgnoreEntities: []string{},
				}

				err := source.Connect()

				if err != nil {
					return nil, err
				}

				return source, nil
			}

			return nil, errors.New("unknown source")
		},
		BackupFunc: func(src cli.Source) {
			cli.BackupTool(source, "clients", Server{}, cli.BackupOpts{
				RenameTo:  "servers",
				IndexCols: []string{},
			})

			cli.BackupTool(source, "users", User{}, cli.BackupOpts{
				IndexCols:  []string{},
				Transforms: userTransforms,
			})

			cli.BackupTool(source, "Widgets", Template{}, cli.BackupOpts{
				RenameTo:  "templates",
				IndexCols: []string{},
			})

			cli.BackupTool(source, "userwidgets", Widget{}, cli.BackupOpts{
				IndexCols:  []string{},
				Transforms: widgetTransforms,
			})

			migrations.Migrate(context.Background(), cli.Pool)
		},
	})
}
