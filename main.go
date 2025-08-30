package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"
)

func main() {
	repo := os.Getenv("GIT_REPOSITORY")
	if repo == "" {
		log.Fatal("GIT_REPOSITORY env must be set")
		return
	}

	wardenClient := os.Getenv("BW_CLIENTID")
	wardenSecret := os.Getenv("BW_CLIENTSECRET")
	wardenPswd := os.Getenv("BW_PASSWORD")

	if wardenClient == "" || wardenSecret == "" || wardenPswd == "" {
		log.Fatal("BW_CLIENTID, BW_CLIENTSECRET and BW_PASSWORD must be set")
		return
	}

	initRepo(repo)

	intervall, isSet := os.LookupEnv("INTERVALL")
	if !isSet {
		intervall = "43200"
	}

	fmt.Println("init Ticker...")
	i, err := strconv.Atoi(intervall)
	if err != nil {
		// ... handle error
		panic(err)
	}
	timer := time.NewTicker(time.Duration(i) * time.Second)

	defer timer.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	doBackup(wardenPswd)

	for {
		select {
		case <-timer.C:
			doBackup(wardenPswd)
		case <-quit:
			fmt.Println("Shutting down...")
			logout := exec.Command("bw", "logout")
			out, err := logout.Output()
			if err != nil {
				fmt.Printf("Error occurred: %v\n", err)         // More details about the error
				fmt.Printf("Command output: %s\n", string(out)) // Output of the command
				log.Fatal(err)
				return
			}
			fmt.Println(string(out))
			return
		}
	}
}

func doBackup(vaultPswd string) {
	fmt.Println("executing backup...")

	unlock := exec.Command("bw", "unlock", "--passwordenv", "BW_PASSWORD")

	out, err := unlock.CombinedOutput()
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err)         // More details about the error
		fmt.Printf("Command output: %s\n", string(out)) // Output of the command
		log.Fatal(err)
		return
	}
	output := string(out)

	re := regexp.MustCompile(`export BW_SESSION="([^"]+)"`)

	// Find the first match
	matches := re.FindStringSubmatch(output)

	export := exec.Command("bw", "export", "--output", "./repo/export.zip", "--format", "zip", "--session", matches[1])

	out, err = export.CombinedOutput()
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err)         // More details about the error
		fmt.Printf("Command output: %s\n", string(out)) // Output of the command
		log.Fatal(err)
	}
	fmt.Println(string(out))

	pushToRepo()

	lock := exec.Command("bw", "lock")

	out, _ = lock.Output()

	fmt.Print(string(out))
}

func initRepo(repo string) {
	fmt.Println("init repo...")

	if _, err := os.Stat("./repo"); errors.Is(err, os.ErrNotExist) {
		cmd := exec.Command("git", "clone", repo, "./repo")
		stdout, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("Error occurred: %v\n", err)            // More details about the error
			fmt.Printf("Command output: %s\n", string(stdout)) // Output of the command
			log.Fatal(err)
			return
		}
		fmt.Println(string(stdout))
	}

	setName := exec.Command("git", "-C", "./repo", "config", "user.name", os.Getenv("GIT_NAME"))

	out, err := setName.CombinedOutput()
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err) // More details about the error
		fmt.Printf("Command output: %s\n", string(out))
		log.Fatal(err.Error())
		return
	}
	fmt.Println(string(out))
	setMail := exec.Command("git", "-C", "./repo", "config", "user.email", os.Getenv("GIT_EMAIL"))

	out, err = setMail.CombinedOutput()
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err) // More details about the error
		fmt.Printf("Command output: %s\n", string(out))
		log.Fatal(err.Error())
		return
	}
	fmt.Println(string(out))

	domain, isSet := os.LookupEnv("VAULT_DOMAIN")
	if isSet {
		fmt.Printf("init bw with domain %s", domain)
		config := exec.Command("bw", "config", "server", domain)
		out, err := config.CombinedOutput()
		if err != nil {
			fmt.Printf("Error occurred: %v\n", err) // More details about the error
			fmt.Printf("Command output: %s\n", string(out))
			log.Fatal(err.Error())
			return
		}
		fmt.Println(string(out))
	}

	fmt.Println("login bw...")
	login := exec.Command("bw", "login", "--apikey")
	out, err = login.CombinedOutput()
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err) // More details about the error
		fmt.Printf("Command output: %s\n", string(out))
		log.Fatal(err.Error())
		return
	}
	fmt.Println(string(out))
}

func pushToRepo() {
	add := exec.Command("git", "-C", "./repo", "add", ".")

	out, err := add.CombinedOutput()
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err) // More details about the error
		fmt.Printf("Command output: %s\n", string(out))
		log.Fatal(err.Error())
		return
	}
	fmt.Println(string(out))

	commit := exec.Command("git", "-C", "./repo", "commit", "-m", "\"backup "+time.Now().Format("200601021504")+"\"")
	out, err = commit.CombinedOutput()
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err) // More details about the error
		fmt.Printf("Command output: %s\n", string(out))
		log.Fatal(err.Error())
		return
	}
	fmt.Println(string(out))

	push := exec.Command("git", "-C", "./repo", "push", "origin", "main")
	out, err = push.CombinedOutput()
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err) // More details about the error
		fmt.Printf("Command output: %s\n", string(out))
		log.Fatal(err.Error())
		return
	}
	fmt.Println(string(out))
}
