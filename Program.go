package main

import (
    "time"
    "os/exec"
    "runtime"
    "fmt"
    "os"
)

func openBackup() error {
    var cmd *exec.Cmd

    // Get the current directory (the one from which the server was run)
    workingDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("could not get current directory: %v", err)
    }

    // Detect the operating system
    switch runtime.GOOS {
    case "linux":
        // Open a new terminal and run the client on Linux (gnome-terminal)
        cmd = exec.Command("gnome-terminal", "--", "go", "run", fmt.Sprintf("cd %s && go run BackupServer.go", workingDir))
    case "darwin":
        // Open a new terminal and run the client on macOS (osascript)
        cmd = exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to do script "cd %s && go run BackupServer.go"`, workingDir))
    case "windows":
        // Open a new terminal and run the client on Windows (start command)
        cmd = exec.Command("cmd", "/c", "start", "cmd", "/k", fmt.Sprintf("cd %s && go run BackupServer.go %s", workingDir))
    default:
        return fmt.Errorf("unsupported platform")
    }

    // Start the command
    return cmd.Start()
}

func openPrimary() error {
    var cmd *exec.Cmd

    // Get the current directory (the one from which the server was run)
    workingDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("could not get current directory: %v", err)
    }

    // Detect the operating system
    switch runtime.GOOS {
    case "linux":
        // Open a new terminal and run the client on Linux (gnome-terminal)
        cmd = exec.Command("gnome-terminal", "--", "go", "run", fmt.Sprintf("cd %s && go run PrimaryServer.go", workingDir))
    case "darwin":
        // Open a new terminal and run the client on macOS (osascript)
        cmd = exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to do script "cd %s && go run PrimaryServer.go"`, workingDir))
    case "windows":
        // Open a new terminal and run the client on Windows (start command)
        cmd = exec.Command("cmd", "/c", "start", "cmd", "/k", fmt.Sprintf("cd %s && go run PrimaryServer.go %s", workingDir))
    default:
        return fmt.Errorf("unsupported platform")
    }

    // Start the command
    return cmd.Start()
}

func openClient(clientID int32) error {
    var cmd *exec.Cmd

    // Get the current directory (the one from which the server was run)
    workingDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("could not get current directory: %v", err)
    }

    // Detect the operating system
    switch runtime.GOOS {
    case "linux":
        // Open a new terminal and run the client on Linux (gnome-terminal)
        cmd = exec.Command("gnome-terminal", "--", "go", "run", fmt.Sprintf("cd %s && go run Client.go %d", workingDir, clientID))
    case "darwin":
        // Open a new terminal and run the client on macOS (osascript)
        cmd = exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to do script "cd %s && go run Client.go %d"`, workingDir, clientID))
    case "windows":
        // Open a new terminal and run the client on Windows (start command)
        cmd = exec.Command("cmd", "/c", "start", "cmd", "/k", fmt.Sprintf("cd %s && go run Client.go %d", workingDir, clientID))
    default:
        return fmt.Errorf("unsupported platform")
    }

    // Start the command
    return cmd.Start()
}

func main() {
	// Simulate opening BackupServer
	err := openBackup()
	if err != nil {
		fmt.Printf("Error opening backup server: %v\n", err)
	}
	time.Sleep(time.Duration(5*time.Second))

	// Simulate opening PrimaryServer
	err = openPrimary()
	if err != nil {
		fmt.Printf("Error opening primary server: %v\n", err)
	}
	time.Sleep(time.Duration(5*time.Second))

	// Simulate opening multiple clients
    for i := 0; i < 4; i++ { // Open 4 clients
        err := openClient(int32(i+1))
        if err != nil {
            fmt.Printf("Error opening client: %v\n", err)
        }
    }
}