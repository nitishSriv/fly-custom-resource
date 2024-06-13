package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type FlyResource struct {
	Source struct {
		Path string `json:"path"`
		Url  string `json:"url"`
	} `json:"source"`
	Version struct {
		Cli      string `json:"cli"`
		Platform string `json:"platform"`
	} `json:"version"`
}

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

func (f *FlyResource) ReadFromStdin() error {
	// Input example
	// scanner := `{
	// 	"source": {
	// 		"path": "/usr/local/bin",
	//      	"url": "https://github.com/concourse/concourse/releases/download/v7.9.1/fly-7.9.1-linux-amd64.tgz"
	// 	},
	// 	"version": {
	// 		"cli": "7.9.1",
	// 		"platform": "linux-amd64"
	// 	}
	// }`

	err := json.NewDecoder(os.Stdin).Decode(&f)
	if err != nil {
		return err
	}
	log.Println(f)

	return nil
}

func ExecCmd(cmdStr string) (string, string, error) {
	cmd := exec.Command("sh", "-c", cmdStr)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Println("[Cmd]: ", cmdStr)

	err := cmd.Run()
	if err != nil {
		fmt.Println("Unable to execute the command", stderr.String())
		return "", "", err
	}

	return stdout.String(), stderr.String(), err
}

func GetVersion(path string) (string, error) {
	cmdStr := fmt.Sprintf("%s/fly --version", path)
	stdout, stderr, err := ExecCmd(cmdStr)
	if err != nil {
		fmt.Println("Error while running the get version command", stderr)
		return "", err
	}

	resVersion := strings.Trim(stdout, "\n")

	return resVersion, nil
}

func (f *FlyResource) InstallFlyCli() (string, error) {
	destDir := os.Args[2]
	if destDir == "" {
		destDir = "downloads"
	}
	flyReleaseUrl := f.Source.Url

	ExecCmd("mkdir " + destDir)

	downloadCmd := fmt.Sprintf("wget %s -O %s/fly-cli-%s.tgz", flyReleaseUrl, destDir, f.Version.Platform)

	_, _, err := ExecCmd(downloadCmd)
	if err != nil {
		log.Println("Error while downloading fly cli: " + err.Error())
		return "", err
	}

	untarCmd := fmt.Sprintf("tar -xvf %s/fly-*.tgz -C %s/", destDir, destDir)
	_, _, err = ExecCmd(untarCmd)
	if err != nil {
		log.Println("Error while extracting fly cli tar: " + err.Error())
		return "", err
	}

	// installCmd := fmt.Sprintf("chmod +x ./%s/fly && sudo mv ./%s/fly /usr/local/bin/", destDir, destDir)
	// _, _, err = ExecCmd(installCmd)
	// if err != nil {
	// 	log.Println("Error while installing fly cli: " + err.Error())
	// 	return err
	// }

	installCmd := fmt.Sprintf("chmod +x ./%s/fly", destDir)
	_, _, err = ExecCmd(installCmd)
	if err != nil {
		log.Println("Error while installing fly cli: " + err.Error())
		return "", err
	}

	cliVersion, err := GetVersion(destDir + "/")
	if err != nil {
		log.Println("Error occurred while getting cli version: " + err.Error())
		return cliVersion, err
	}

	log.Println("Installed fly cli version : " + cliVersion)

	return cliVersion, nil
}
