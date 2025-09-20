package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Runner struct {
	Request             *plugin.CodeGeneratorRequest
	Response            *plugin.CodeGeneratorResponse
	modes               []string // "xml" or "json"
	pathsSourceRelative bool
}

func parsePackageOption(file *descriptorpb.FileDescriptorProto) (packagePath string, pkg string, ok bool) {
	opt := file.GetOptions().GetGoPackage()
	if opt == "" {
		return "", "", false
	}
	sc := strings.Index(opt, ";")
	if sc >= 0 {
		return opt[:sc], opt[sc+1:], true
	}
	slash := strings.LastIndex(opt, "/")
	if slash >= 0 {
		return opt, opt[slash+1:], true
	}
	return "", opt, true
}

func (runner *Runner) parseParams(param string) error {
	if strings.TrimSpace(param) == "" {
		return fmt.Errorf("no parameter provided: expected 'xml', 'json' or 'gql")
	}
	for _, part := range strings.Split(param, ",") {
		part = strings.TrimSpace(part)
		switch {
		case part == "xml" || part == "json" || part == "gql":
			runner.modes = append(runner.modes, part)
			continue
		case part == "paths=source_relative":
			runner.pathsSourceRelative = true
		default:
		}
	}
	if runner.modes == nil {
		return fmt.Errorf("unknown or missing parameter: got %q, want 'xml, 'json' or 'gql'", param)
	}
	return nil
}

func (runner *Runner) getFileName(file *descriptorpb.FileDescriptorProto, mode string) (string, error) {
	name := file.GetName()
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}

	switch mode {
	case "xml":
		name += ".enums.xml.go"
	case "json":
		name += ".enums.json.go"
	case "gql":
		name += ".enums.gql.go"
	default:
		return "", fmt.Errorf("unsupported mode %q", mode)
	}

	// If paths=source_relative is NOT set, mirror the original behavior:
	// move the file under the go_package import path when present.
	if !runner.pathsSourceRelative {
		if packagePath, _, ok := parsePackageOption(file); ok && packagePath != "" {
			_, base := path.Split(name)
			name = path.Join(packagePath, base)
		}
	}

	return name, nil
}

func (runner *Runner) generateMarshallers(fileTemplate *template.Template, enumTemplate *template.Template, mode string) error {
	for _, file := range runner.Request.ProtoFile {
		fileContent, err, found := applyTemplate(file, fileTemplate, enumTemplate)
		if err != nil {
			return err
		}
		if found {
			filename, err := runner.getFileName(file, mode)
			if err != nil {
				return err
			}
			outFile := &plugin.CodeGeneratorResponse_File{
				Name:    &filename,
				Content: &fileContent,
			}
			runner.Response.File = append(runner.Response.File, outFile)
		}
	}
	return nil
}

func (runner *Runner) generateCode() error {
	runner.Response.File = make([]*plugin.CodeGeneratorResponse_File, 0)

	// Parse params once
	if err := runner.parseParams(runner.Request.GetParameter()); err != nil {
		return err
	}

	var err error

	for _, mode := range runner.modes {
		switch mode {
		case "xml":
			err = runner.generateMarshallers(xmlFileTemplate, xmlEnumTemplate, mode)
		case "json":
			err = runner.generateMarshallers(jsonFileTemplate, jsonEnumTemplate, mode)
		case "gql":
			err = runner.generateMarshallers(gqlFileTemplate, gqlEnumTemplate, mode)
		default:
			err = fmt.Errorf("unknown mode %q", mode)
		}
	}

	return err
}

var SupportedFeatures = uint64(plugin.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

func main() {
	req := &plugin.CodeGeneratorRequest{}
	resp := &plugin.CodeGeneratorResponse{
		SupportedFeatures: &SupportedFeatures,
	}

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	// You must use the requests unmarshal method to handle this type
	if err := proto.Unmarshal(data, req); err != nil {
		panic(err)
	}

	runner := &Runner{
		Request:  req,
		Response: resp,
	}

	err = runner.generateCode()
	if err != nil {
		panic(err)
	}

	marshalled, err := proto.Marshal(resp)
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(marshalled)
}
