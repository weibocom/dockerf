package image

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const TMP_DIR = "/tmp/ForBuildDockerImage/"

func run(dir, cmdline string, args ...string) ([]byte, error) {
	_, err := exec.LookPath(cmdline)
	if err != nil {
		fmt.Printf("Not install '%s'\n", cmdline)
		return nil, err
	}

	cmd := exec.Command(cmdline, args...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	out := buf.Bytes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "# cd %s; %s %s\n", dir, cmd, strings.Join(args, " "))
		os.Stderr.Write(out)
		return nil, err
	}
	return out, nil
}

func gitCloneRepoIfNeed(dir, repo string) (string, error) {
	// TODO check whether exist lastest source code from git
	//if _, err := os.Stat(dir); err == nil { // exist

	// rm -rf dir
	_, err := run(".", "rm", "-rf", dir)
	if err != nil {
		fmt.Printf("Failed to rm exist source dir '%s'\n", dir)
		return "", err
	}
	fmt.Printf("Removed exist source dir '%s'\n", dir)

	// mkdir -p dir
	_, err = run(".", "mkdir", "-p", dir)
	if err != nil {
		fmt.Printf("Failed to create source dir '%s'\n", dir)
		return "", err
	}
	fmt.Printf("Create source dir '%s'\n", dir)

	// git clone repo dir
	_, err = run(".", "git", "clone", repo, dir)
	if err != nil {
		fmt.Printf("Failed to git clone source code '%s'\n", repo)
		return "", err
	}

	// get commit id
	out, err := run(dir, "git", "rev-parse", "HEAD")
	if err != nil {
		fmt.Printf("Failed to get git commit id\n")
		return "", err
	}
	commitId := string(bytes.TrimSpace(out))

	return commitId, nil
}

func buildAndPushImageIfNeed(dir, imageURL string) error {
	// TODO check whether exist image in registry

	// docker build -t imageName:tag ./
	_, err := run(dir, "docker", "build", "-t", imageURL, "./")
	if err != nil {
		fmt.Printf("Failed to build docker image\n")
		return err
	}

	// docker push image
	_, err = run(".", "docker", "push", imageURL)
	if err != nil {
		fmt.Printf("Failed to push image '%s'\n", imageURL)
		return err
	}

	return nil
}

// @return: append commitId for building imageURL
func BuildAndPushByRepo(repo, image string) (string, error) {
	localSource := TMP_DIR + image
	commitId, err := gitCloneRepoIfNeed(localSource, repo)
	if err != nil {
		return "", err
	}
	fmt.Printf("Git clone source code '%s'\n", repo)

	imageURL := image + ":" + commitId
	err = buildAndPushImageIfNeed(localSource, imageURL)
	if err != nil {
		return "", err
	}
	fmt.Printf("Build and push to Registry OK\n")

	return imageURL, nil
}
