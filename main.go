package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"google.golang.org/protobuf/types/descriptorpb"
)

type XMLEnums struct {
	Request  *plugin.CodeGeneratorRequest
	Response *plugin.CodeGeneratorResponse
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

func (runner *XMLEnums) getFileName(file *descriptorpb.FileDescriptorProto) (string, error) {
	name := *file.Name
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}

	parameter := runner.Request.GetParameter()
	if parameter == "xml" {
		name += ".xml.go"
	} else if parameter == "json" {
		name += ".json.go"
	} else {
		return "", fmt.Errorf("Unknown parameter %s", parameter)
	}

	if packagePath, _, ok := parsePackageOption(file); ok && packagePath != "" {
		_, name = path.Split(name)
		name = path.Join(string(packagePath), name)
	}

	return name, nil
}

func (runner *XMLEnums) generateMarshallers(fileTemplate *template.Template, enumTemplate *template.Template) error {

	for _, file := range runner.Request.ProtoFile {
		fileContent, err, found := applyTemplate(file, fileTemplate, enumTemplate)

		if err != nil {
			return err
		}

		if found {
			filename, err := runner.getFileName(file)

			if err != nil {
				return err
			}

			var outFile plugin.CodeGeneratorResponse_File
			outFile.Name = &filename
			outFile.Content = &fileContent

			runner.Response.File = append(runner.Response.File, &outFile)
		}
	}

	return nil
}

func (runner *XMLEnums) generateCode() error {
	// Initialize the output file slice
	files := make([]*plugin.CodeGeneratorResponse_File, 0)
	runner.Response.File = files

	var err error
	parameter := runner.Request.GetParameter()
	if parameter == "xml" {
		err = runner.generateMarshallers(xmlFileTemplate, xmlEnumTemplate)
	} else if parameter == "json" {
		err = runner.generateMarshallers(jsonFileTemplate, jsonEnumTemplate)
	} else {
		return fmt.Errorf("Unknown parameter %s", parameter)
	}

	if err != nil {
		return err
	}

	return nil
}

func main() {
	req := &plugin.CodeGeneratorRequest{}
	resp := &plugin.CodeGeneratorResponse{}

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	// You must use the requests unmarshal method to handle this type
	if err := proto.Unmarshal(data, req); err != nil {
		panic(err)
	}

	runner := &XMLEnums{
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
