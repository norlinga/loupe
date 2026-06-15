package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	loupecontext "github.com/norlinga/loupe/internal/context"
	"github.com/norlinga/loupe/internal/observe"
	"github.com/norlinga/loupe/internal/schema"
	"github.com/norlinga/loupe/internal/version"
)

func Run(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("loupe", flag.ContinueOnError)
	fs.SetOutput(stderr)
	depth := fs.Int("depth", -1, "recurse N levels deep")
	typeFilter := fs.String("type", "", "filter entries: file, dir, symlink")
	newerThan := fs.Int64("newer-than", 0, "only entries modified in the last N seconds")
	noHidden := fs.Bool("no-hidden", false, "exclude hidden entries")
	withContext := fs.Bool("context", false, "add semantic enrichment layer")
	human := fs.Bool("human", false, "human-readable output")
	showVersion := fs.Bool("version", false, "print version")
	flagArgs, path, err := splitArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return err
	}
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if *showVersion {
		fmt.Fprintln(stdout, version.String())
		return nil
	}
	if path == "" {
		err := errors.New("usage: loupe <path> [flags]")
		fmt.Fprintln(stderr, err)
		return err
	}
	parsedType, err := parseType(*typeFilter)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return err
	}
	node, err := observe.Observe(path, observe.Options{
		Depth:     *depth,
		Type:      parsedType,
		NewerThan: time.Duration(*newerThan) * time.Second,
		NoHidden:  *noHidden,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return err
	}
	if *withContext {
		loupecontext.Enrich(node, loupecontext.Options{})
	}
	if *human {
		writeHuman(stdout, *node, 0)
		return nil
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(node)
}

func splitArgs(args []string) ([]string, string, error) {
	var flagArgs []string
	var path string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			if i+1 >= len(args) || path != "" {
				return nil, "", errors.New("usage: loupe <path> [flags]")
			}
			path = args[i+1]
			if i+2 != len(args) {
				return nil, "", errors.New("usage: loupe <path> [flags]")
			}
			break
		}
		if strings.HasPrefix(arg, "-") && arg != "-" {
			flagArgs = append(flagArgs, arg)
			if flagNeedsValue(arg) && !strings.Contains(arg, "=") {
				i++
				if i >= len(args) {
					return nil, "", fmt.Errorf("flag needs an argument: %s", arg)
				}
				flagArgs = append(flagArgs, args[i])
			}
			continue
		}
		if path != "" {
			return nil, "", errors.New("usage: loupe <path> [flags]")
		}
		path = arg
	}
	return flagArgs, path, nil
}

func flagNeedsValue(arg string) bool {
	name := strings.TrimLeft(arg, "-")
	if before, _, ok := strings.Cut(name, "="); ok {
		name = before
	}
	switch name {
	case "depth", "type", "newer-than":
		return true
	default:
		return false
	}
}

func parseType(value string) (schema.EntryType, error) {
	switch strings.ToLower(value) {
	case "":
		return "", nil
	case "file":
		return schema.TypeFile, nil
	case "dir", "directory":
		return schema.TypeDirectory, nil
	case "symlink":
		return schema.TypeSymlink, nil
	default:
		return "", fmt.Errorf("unsupported --type %q", value)
	}
}

func writeHuman(w io.Writer, node schema.Node, indent int) {
	prefix := strings.Repeat("  ", indent)
	fmt.Fprintf(w, "%s%s %s %d bytes %s\n", prefix, node.Type, node.Path, node.SizeBytes, node.Permissions)
	for _, entry := range node.Entries {
		writeHuman(w, entry, indent+1)
	}
}
