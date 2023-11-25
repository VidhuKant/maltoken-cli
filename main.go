/*
maltoken-cli - MyAnimeList auth token generator
Copyright © 2023 Vidhu Kant Sharma <vidhukant@vidhukant.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	mt "vidhukant.com/maltoken"
)

const VERSION = "v1.0.0"

var (
	cid string
	port int
	launch bool
)

var HTMLTemplate = `
<html>
  <head>
    <title>maltoken-cli</title>
  </head>
  <body>
    <style>
    body {
      background-color: #232627;
      display: flex;
      flex-direction: column;
      justify-content: center;
      gap: 0.5rem;
      align-items: center;
      min-height: 100vh;
    }
    #heading, #subheading, #description {
      text-align: center;
      margin: 0;
    }
    #heading {
      color: #C678DD;
      font-size: 2.2em;
    }
    #subheading {
      color: #dfdfdf;
      font-size: 1.2em;
    }
    #description {
      color: lightgray;
      font-size: 0.9em;
    }
    </style>
    <p id="heading">%s</p>
    <p id="subheading">%s</p>
    <p id="description">maltoken-cli version ` + VERSION + `</p>
  </body>
</html>
`

func main() {
	if strings.TrimSpace(cid) == "" {
		scanner := bufio.NewScanner(os.Stdin)

		fmt.Printf("Enter your Client ID: ")
		scanner.Scan()

		if scanner.Err() != nil {
			fmt.Printf("\x1b[1;31mAn error occoured while reading Client ID:\x1b[0m %s\n", scanner.Err().Error())
		}

		if strings.TrimSpace(scanner.Text()) == "" {
			fmt.Println("\x1b[1;31mInvalid Client ID.\x1b[0m")
			os.Exit(1)
		}

		cid = scanner.Text()
	}

	challenge, link := mt.GetChallengeLink(cid)

	fmt.Printf("Authorization URL: \x1b[36m%s\x1b[0m\n", link)

	if launch {
		fmt.Println("Attempting to launch the browser...")

		var err error
		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", link).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", link).Start()
		case "darwin":
			err = exec.Command("open", link).Start()
		default:
			err = errors.New("Failed to detect platform.")
		}

		if err != nil {
			fmt.Printf("\x1b[1;31mFailed to launch the browser due to the following error:\x1b[0m %s\n", err.Error())
			fmt.Println("Please manually copy and paste the link.")
		}
	}

	res, err := mt.Listen(cid, challenge, port)
	if err != nil {
		fmt.Printf("\x1b[1;31mAn error occoured:\x1b[0m %s\nExiting...\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("\x1b[1;33mToken Type:\x1b[0m %v\n", res["token_type"])
	fmt.Printf("\x1b[1;33mExpires In:\x1b[0m %v\n", res["expires_in"])
	fmt.Printf("\x1b[1;33mAccess Token:\x1b[0m \x1b[36m%v\x1b[0m\n", res["access_token"])
	fmt.Printf("\x1b[1;33mRefresh Token:\x1b[0m \x1b[36m%v\x1b[0m\n", res["refresh_token"])
}

func version(_ string) error {
	fmt.Printf("maltoken-cli version %s %s/%s\n", VERSION, runtime.GOOS, runtime.GOARCH)
	os.Exit(0)
	return nil
}

func help(_ string) error {
	fmt.Println("maltoken-cli: MyAnimeList auth token generator")

	fmt.Println("\n\x1b[34mmaltoken-cli  Copyright (C) 2023  Vidhu Kant Sharma <vidhukant@vidhukant.com>\n" +
		"This program comes with ABSOLUTELY NO WARRANTY;\n" +
		"This is free software, and you are welcome to redistribute it\n" +
		"under certain conditions; For details refer to the GNU General Public License.\n" +
		"You should have received a copy of the GNU General Public License\n" +
		"along with this program.  If not, see <\x1b[36mhttps://www.gnu.org/licenses/\x1b[34m>.\x1b[0m\n",
	)

	fmt.Println("Usage:")
	fmt.Println("  maltoken-cli [flags]")

	fmt.Println("Flags:")
	fmt.Println("  --client-id  \t Specify the Client ID")
	fmt.Println("  --port       \t Specify the port to run the server on (default 8080)")
	fmt.Println("  --launch     \t Launch authorization page in the browser automatically (might not work on some systems)")
	fmt.Println("  -v, --version\t Print version number")
	fmt.Println("  -h, --help   \t Show this message")

	fmt.Println("\nCheck out maltoken <\x1b[36mhttps://mikunonaka.net/maltoken/about\x1b[0m> to embed this functionality in your go project.")

	os.Exit(0)
	return nil
}

func init() {
	flag.Usage = func() {
		help("")
	}

	flag.IntVar(&port, "port", 8080, "Specify the port to run the server on")
	flag.StringVar(&cid, "client-id", "", "Specify the Client ID")
	flag.BoolVar(&launch, "launch", false, "Launch authorization page in the browser automatically (might not work on some systems)")

	flag.BoolFunc("version", "Print the version number", version)
	flag.BoolFunc("v", "Print the version number", version)

	flag.Parse()

	mt.SuccessHTML = fmt.Sprintf(HTMLTemplate, "Yay! Authorization Successful.", "You may close this tab now.")
	mt.BadRequestHTML = fmt.Sprintf(HTMLTemplate, "Invalid request.", "Required query parameters are missing.")
	mt.ErrHTML = fmt.Sprintf(HTMLTemplate, "An error occoured.", "%s")
}
