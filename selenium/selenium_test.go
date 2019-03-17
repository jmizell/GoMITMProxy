// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.
// +build integration_test

package selenium

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/tebeka/selenium"

	"github.com/jmizell/GoMITMProxy/proxy/config"
	"github.com/jmizell/GoMITMProxy/proxy/log"
)

func killAndRemoveContainer() {

	cmd := exec.Command("docker", "stop", "selenium_integration_test")
	_, _ = cmd.CombinedOutput()

	cmd = exec.Command("docker", "rm", "selenium_integration_test")
	_, _ = cmd.CombinedOutput()
}

func TestSelenium(t *testing.T) {

	/*
	START THE WEB HOOK SERVER
	 */
	receivedMessages := make([]*log.MSG, 0)
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
	}
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

		if request.Method != http.MethodPost {
			fmt.Printf("SERVER: incoming request should be %s, but received %s\n", http.MethodPost, request.Method)
			t.Fail()
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var msg log.MSG
		//err := json.NewDecoder(request.Body).Decode(&msg)
		data, err := ioutil.ReadAll(request.Body)
		if err != nil {
			fmt.Printf("SERVER: failed to read webhook payload to message, %s\n", err.Error())
			t.Fail()
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(data, &msg)
		if err != nil {
			fmt.Printf("SERVER: %s\n", string(data))
			fmt.Printf("SERVER: failed to decode webhook payload to message, %s\n", err.Error())
			t.Fail()
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		receivedMessages = append(receivedMessages, &msg)
		writer.WriteHeader(http.StatusOK)
	})

	connection, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("failed to open a port of for the test server, %s", err.Error())
	}
	port := connection.Addr().(*net.TCPAddr).Port

	go server.Serve(connection)
	defer server.Shutdown(context.Background())
	fmt.Printf("SERVER: server listen address 127.0.0.1:%d\n", port)

	/*
	WRITE THE CONFIG FILE
	 */
	cfg := &config.Config{
		CAKeyFile:           "/etc/gomitmproxy/ca.key",
		CACertFile:          "/etc/gomitmproxy/ca.crt",
		HTTPPorts:           []int{80, 8080, 8880, 2052, 2082, 2086, 2095},
		HTTPSPorts:          []int{443, 2053, 2083, 2087, 2096, 8443},
		ListenAddr:          "127.0.0.50",
		DNSPort:             53,
		DNSRegex:            ".*",
		DNSResolverOverride: "8.8.8.8",
		ForwardDNSServer:    "8.8.8.8",
		LogResponses:        true,
		Level:               log.DEBUG,
		Format:              log.TEXT,
		WebHookURL:          fmt.Sprintf("http://127.0.0.1:%d", port),
	}
	cfgBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to serialize config, %s", err.Error())
	}
	fmt.Println("DEBUG -- -- ", string(cfgBytes))
	configFile, err := ioutil.TempFile("", "*.json")
	if _, err = configFile.Write(cfgBytes); err != nil {
		t.Fatalf("failed to write config, %s", err.Error())
	}
	if err = configFile.Close(); err != nil {
		t.Fatalf("failed to close config file, %s", err.Error())
	}

	/*
	START THE SELENIUM CONTAINER
	 */
	killAndRemoveContainer()
 	cmd := exec.Command(
 		"docker",
 		"run",
 		"-i",
 		"--rm",
 		"--name", "selenium_integration_test",
 		"-v", fmt.Sprintf("%s:/etc/gomitmproxy/config.json", configFile.Name()),
 		"--net", "host",
 		"jmizell/gomitmproxy:selenium_integration_test")


	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to create command stdout pipe, %s", err.Error())
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("failed to create command stderr pipe, %s", err.Error())
	}

	go func() {
		reader := bufio.NewReader(stdout)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			fmt.Printf("DOCKER: %s\n", scanner.Text())
		}
	}()

	go func() {
		reader := bufio.NewReader(stderr)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			fmt.Printf("DOCKER: %s\n", scanner.Text())
		}
	}()

	if 	err := cmd.Start(); err != nil {
		t.Fatalf("failed to start container, %s", err.Error())
	}
	defer func() {
		killAndRemoveContainer()

	}()

 	/*
	WAIT FOR CONTAINER TO BE READY
 	*/
	caps := selenium.Capabilities{"browserName": "chrome"}
	var wd selenium.WebDriver
	var portUp bool
 	for i := 0; i < 150; i++ {
		wd, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 4444))
		if err == nil {
			portUp = true
			break
		}
		time.Sleep(time.Millisecond * 200)
	}

 	if !portUp {
 		t.Logf("timed out waiting for docker container to be ready")
 		t.Fail()
	}

 	if wd != nil {
		defer wd.Quit()
	}

 	/*
 	RUN THE TESTS
 	 */
	if err := wd.Get("https://www.github.com"); err != nil {
		t.Logf("failed to get website, %s", err.Error())
		t.Fail()
	}

	if title, err := wd.Title(); err != nil {
		t.Logf("failed to get title from browser %s", err.Error())
		t.Fail()
	} else if strings.Contains(title, "Github") {
		t.Logf("Github not in title, %s", title)
		t.Fail()
	}

	screenShot, err := wd.Screenshot()
	if err != nil {
		t.Logf("failed to get screenshot, %s", err.Error())
		t.Fail()
	}

	if _, err := os.Stat("/tmp/selenium_screenshot.png"); !os.IsNotExist(err) {
		err = os.Remove("/tmp/selenium_screenshot.png")
		if err != nil {
			t.Fatalf("failed to remove old screenshot, %s", err.Error())
		}
	}

	f, err := os.OpenFile("/tmp/selenium_screenshot.png", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		t.Logf("failed to open screenshot file for writing, %s", err.Error())
		t.Fail()
	}
	defer f.Close()

	_, err = f.Write(screenShot)
	if err != nil {
		t.Logf("failed to write screenshot, %s", err.Error())
		t.Fail()
	}

	time.Sleep(time.Second * 5)

	/*
	INSPECT THE WEB HOOK LOG
	 */
 	fmt.Println("DEBUG -- -- len(receivedMessages)", len(receivedMessages))
 	if len(receivedMessages) == 0 {
 		t.Logf("expected more then zero webhook messages, found %d", len(receivedMessages))
 		t.Fail()
	}

 	var foundGithub bool
 	var foundGithubAssets bool
 	for _, v := range receivedMessages {
 		if v.Request != nil {
 			if strings.Contains(v.Request.URL.String(), "www.github.com") {
 				foundGithub = true
			}
			if strings.Contains(v.Request.URL.String(), "github.githubassets.com") {
				foundGithubAssets = true
			}
 			fmt.Printf("RECEIVED: %s %s\n", v.Request.Method, v.Request.URL.String())
 		}
	}

 	if !foundGithub {
 		t.Logf("could not find www.github.com")
 		t.Fail()
	}

	if !foundGithubAssets {
		t.Logf("could not find github.githubassets.com")
		t.Fail()
	}
}
