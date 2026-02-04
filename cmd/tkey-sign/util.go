// SPDX-FileCopyrightText: 2023 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"

	"github.com/tillitis/tkey-sign-cli/signify"
	"github.com/tillitis/tkeyclient"
	"github.com/tillitis/tkeysign"
)

const (
	// 4 chars each.
	wantFWName0  = "tk1 "
	wantFWName1  = "mkdf"
	wantAppName0 = "tk1 "
	wantAppName1 = "sign"
)

func isFirmwareMode(tk *tkeyclient.TillitisKey) bool {
	nameVer, err := tk.GetNameVersion()
	if err != nil {
		return false
	}
	// not caring about nameVer.Version
	return nameVer.Name0 == wantFWName0 &&
		nameVer.Name1 == wantFWName1
}

func isWantedApp(signer tkeysign.Signer) bool {
	nameVer, err := signer.GetAppNameVersion()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			le.Printf("GetAppNameVersion: %s\n", err)
		}
		return false
	}

	// not caring about nameVer.Version
	return nameVer.Name0 == wantAppName0 &&
		nameVer.Name1 == wantAppName1
}

func handleSignals(action func(), sig ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sig...)
	go func() {
		for {
			<-ch
			action()
		}
	}()
}

func readBuildInfo() string {
	var v string

	if info, ok := debug.ReadBuildInfo(); ok {
		sb := strings.Builder{}
		sb.WriteString("devel")
		for _, setting := range info.Settings {
			if strings.HasPrefix(setting.Key, "vcs") {
				sb.WriteString(fmt.Sprintf(" %s=%s", setting.Key, setting.Value))
			}
		}
		v = sb.String()
	}
	return v
}

// writeRetry writes the data in the file given in filename as base64.
// If a file already exists it prompts interactively for permission to
// overwrite the file.
func writeRetry(filename string, data signify.Data, comment string) error {
	err := data.ToFile(filename, comment, false)
	if os.IsExist(errors.Unwrap(err)) {
		le.Printf("File %v exists. Overwrite [y/n]?", filename)
		reader := bufio.NewReader(os.Stdin)
		overWriteP, _ := reader.ReadString('\n')

		// Trim space to normalize Windows line endings
		overWriteP = strings.TrimSpace(overWriteP)

		if overWriteP == "y" {
			err = data.ToFile(filename, comment, true)
		} else {
			le.Printf("Aborted\n")
			os.Exit(1)
		}
	}

	if !os.IsExist(errors.Unwrap(err)) && err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
