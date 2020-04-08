package main

import (
	"log"

	"github.com/mewkiz/pkg/imgutil"
	"github.com/mewspring/ren/pkg/assets"
	"github.com/pkg/errors"
)

func main() {
	if err := ren(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func ren() error {
	area, err := assets.LoadArea("yenwood")
	if err != nil {
		return errors.WithStack(err)
	}
	if err := imgutil.WriteFile("foo.png", area.BackgroundLayer); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
