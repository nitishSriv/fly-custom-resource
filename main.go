package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

type Version struct {
	Ref string `json:"ref"`
}

type CheckResponse []Version

type InResponse struct {
	Version  Version             `json:"version"`
	Metadata []map[string]string `json:"metadata"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("expected command")
	}

	var flyResource FlyResource

	err := flyResource.ReadFromStdin()
	if err != nil {
		log.Fatal(err)
	}

	_, err = flyResource.InstallFlyCli()
	if err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "check":
		err := check(flyResource)
		if err != nil {
			log.Fatal(err)
		}
	case "in":
		err := in(flyResource)
		if err != nil {
			log.Fatal(err)
		}
	case "out":
		err := out(flyResource)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}

func check(flyResource FlyResource) error {
	currVersion, err := GetVersion(flyResource.Source.Path)
	if err != nil {
		log.Println("Error occurred while getting the cli version: " + err.Error())
		return err
	}

	log.Println("Current CLI version : " + currVersion)
	currVersionNum, _ := strconv.ParseFloat(currVersion, 64)

	givenVersion, _ := strconv.ParseFloat(flyResource.Version.Cli, 64)

	response := CheckResponse{
		Version{Ref: flyResource.Version.Cli},
	}

	if currVersionNum != givenVersion {
		log.Println("Version change detected")
		response = append(response, Version{Ref: currVersion})
	}

	return json.NewEncoder(os.Stdout).Encode(response)
}

func in(flyResource FlyResource) error {
	installedVersion, err := flyResource.InstallFlyCli()
	if err != nil {
		log.Fatal(err)
	}

	response := InResponse{
		Version: Version{Ref: installedVersion},
		Metadata: []map[string]string{
			{"name": "fly_cli", "platform": flyResource.Version.Platform},
		},
	}

	return json.NewEncoder(os.Stdout).Encode(response)
}

func out(flyResource FlyResource) error {
	currVersion, err := GetVersion(flyResource.Source.Path)
	if err != nil {
		log.Println("Error occurred while getting the cli version: " + err.Error())
		return err
	}

	response := InResponse{
		Version: Version{Ref: currVersion},
		Metadata: []map[string]string{
			{"name": "fly_cli", "platform": flyResource.Version.Platform},
		},
	}

	return json.NewEncoder(os.Stdout).Encode(response)

}
