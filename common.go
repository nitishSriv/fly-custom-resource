package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
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

func (f *FlyResource) ReadFromStdin() error {
	// scanner := bufio.NewScanner(os.Stdin)
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

	// for scanner.Scan() {
	// 	fmt.Println(scanner.Text())
	// }

	// if scanner.Err() != nil {
	// 	panic("Error occured while reading stdin")
	// }

	err := json.NewDecoder(os.Stdin).Decode(&f)
	if err != nil {
		return err
	}

	// err := json.Unmarshal([]byte(scanner.Text()), &f)
	// // err := json.Unmarshal([]byte(scanner), &f)
	// if err != nil {
	// 	panic("Error while converting the stdin to json: " + err.Error())
	// }
	fmt.Println(f)

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

	installCmd := fmt.Sprintf("chmod +x ./%s/fly", destDir, destDir)
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
