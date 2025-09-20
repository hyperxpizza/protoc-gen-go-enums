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
	modes               []string
	pathsSourceRelative bool
	seen                map[string]struct{}
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
		return fmt.Errorf("no parameter provided: expected 'xml', 'json' or 'gql'")
	}
	for _, part := range strings.Split(param, ",") {
		part = strings.TrimSpace(part)
		switch {
		case part == "xml" || part == "json" || part == "gql":
			runner.modes = append(runner.modes, part)
		case part == "paths=source_relative":
			runner.pathsSourceRelative = true
		case part == "":
		default:
			return fmt.Errorf("unknown parameter %q (allowed: 'xml', 'json', 'gql', 'paths=source_relative')", part)
		}
	}
	if len(runner.modes) == 0 {
		return fmt.Errorf("missing mode: expected one or more of 'xml', 'json', 'gql'")
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
	if !runner.pathsSourceRelative {
		if packagePath, _, ok := parsePackageOption(file); ok && packagePath != "" {
			_, base := path.Split(name)
			name = path.Join(packagePath, base)
		}
	}
	return name, nil
}

func (runner *Runner) filesToGenerate() ([]*descriptorpb.FileDescriptorProto, error) {
	byName := make(map[string]*descriptorpb.FileDescriptorProto, len(runner.Request.ProtoFile))
	for _, fd := range runner.Request.ProtoFile {
		byName[fd.GetName()] = fd
	}
	out := make([]*descriptorpb.FileDescriptorProto, 0, len(runner.Request.FileToGenerate))
	for _, name := range runner.Request.FileToGenerate {
		fd, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("file_to_generate %q missing from request descriptors", name)
		}
		out = append(out, fd)
	}
	return out, nil
}

func (runner *Runner) generateMarshallers(fileTemplate *template.Template, enumTemplate *template.Template, mode string) error {
	targets, err := runner.filesToGenerate()
	if err != nil {
		return err
	}
	for _, file := range targets {
		fileContent, err, found := applyTemplate(file, fileTemplate, enumTemplate)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		filename, err := runner.getFileName(file, mode)
		if err != nil {
			return err
		}
		if _, exists := runner.seen[filename]; exists {
			continue
		}
		runner.seen[filename] = struct{}{}
		outFile := &plugin.CodeGeneratorResponse_File{
			Name:    &filename,
			Content: &fileContent,
		}
		runner.Response.File = append(runner.Response.File, outFile)
	}
	return nil
}

func (runner *Runner) generateCode() error {
	runner.Response.File = make([]*plugin.CodeGeneratorResponse_File, 0)
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
		if err != nil {
			return err
		}
	}
	return nil
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
	if err := proto.Unmarshal(data, req); err != nil {
		panic(err)
	}
	runner := &Runner{
		Request:  req,
		Response: resp,
		seen:     make(map[string]struct{}),
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
