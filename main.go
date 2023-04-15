package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/osquery/osquery-go"
	"github.com/osquery/osquery-go/plugin/table"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf(`Usage: %s SOCKET_PATH`, os.Args[0])
	}

	server, err := osquery.NewExtensionManagerServer("extable", os.Args[1])
	if err != nil {
		log.Fatalf("Error creating extension: %s\n", err)
	}

	server.RegisterPlugin(table.NewPlugin("extable", ExecColumns(), ExecGenerate))
	if err := server.Run(); err != nil {
		log.Fatalln(err)
	}
}

func ExecColumns() []table.ColumnDefinition {
	return []table.ColumnDefinition{
		table.TextColumn("cmd"),
		table.TextColumn("stdout"),
		table.TextColumn("stderr"),
		table.TextColumn("code"),
	}
}

func ExecGenerate(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	if cnstList, present := queryContext.Constraints["cmd"]; present {

		for _, cnst := range cnstList.Constraints {
			if cnst.Operator == table.OperatorEquals {
				cmdArr := strings.Split(cnst.Expression, " ")
				out, err, code := execute(cmdArr[0], cmdArr[1:]...)
				return []map[string]string{
					{
						"cmd":    cnst.Expression,
						"stdout": out,
						"stderr": err,
						"code":   strconv.Itoa(code),
					},
				}, nil
			}
		}
	}
	return nil, errors.New("")
}

func execute(bin string, args ...string) (string, string, int) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return strings.Trim(stdout.String(), " \n"), strings.Trim(stderr.String(), " \n"), exitError.ExitCode()
		}
		return "", err.Error(), -1
	}
	return stdout.String(), stderr.String(), 0
}
