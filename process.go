package main

import (
	"fmt"
	"strings"

	pgs "github.com/lyft/protoc-gen-star/v2"
	"google.golang.org/grpc/codes"

	"github.com/menta2k/protoc-gen-redact/v3/redact/v3"
)

const (
	// defaultErrMsg: for the service method/rpc redaction
	defaultErrMsg = `Permission Denied. Method: "%service%.%method%" has been redacted`
	// error message format specifiers
	specifierMethod  = "%method%"
	specifierService = "%service%"
)

// Process processes the file and adds its generated code into Module.Artifacts
func (m *Module) Process(file pgs.File) {
	// Validate file before processing
	if err := m.validateFile(file); err != nil {
		m.Failf("Cannot process file: %v", err)
		return
	}

	// Add panic recovery for robustness
	defer m.recoverFromPanic(fmt.Sprintf("processing file %s", file.Name()))

	// check file option: FileSkip
	fileSkip := false
	m.must(file.Extension(redact.E_FileSkip, &fileSkip))
	if fileSkip {
		m.Debug(fmt.Sprintf("Skipping file %s due to file_skip option", file.Name()))
		return
	}

	// imports and their aliases
	path2Alias, alias2Path := m.importPaths(file)
	nameWithAlias := func(n pgs.Entity) string {
		imp := m.ctx.ImportPath(n).String()
		name := m.ctx.Name(n).String()
		if alias := path2Alias[imp]; alias != "" {
			name = alias + "." + name
		}
		return name
	}

	data := &ProtoFileData{
		Source:     file.Name().String(),
		Package:    m.ctx.PackageName(file).String(),
		Imports:    alias2Path,
		References: m.references(file, nameWithAlias),
		Services:   make([]*ServiceData, 0, len(file.Services())),
		Messages:   make([]*MessageData, 0, len(file.AllMessages())),
	}

	// all services
	for _, srv := range file.Services() {
		data.Services = append(data.Services, m.processService(srv, nameWithAlias))
	}

	// all messages
	for _, msg := range file.AllMessages() {
		data.Messages = append(data.Messages, m.processMessage(msg, nameWithAlias, true))
	}

	// render file in the template
	name := m.ctx.OutputPath(file).SetExt(".redact.go")
	m.AddGeneratorTemplateFile(name.String(), m.tmpl, data)
}

// processService extracts all pgs.Service and their pgs.Method(s) information and
// structures them into ServiceData
func (m *Module) processService(
	srv pgs.Service,
	nameWithAlias func(n pgs.Entity) string,
) *ServiceData {
	// Validate service before processing
	if err := m.validateService(srv); err != nil {
		m.Failf("Cannot process service: %v", err)
		return nil
	}

	defer m.recoverFromPanic(fmt.Sprintf("processing service %s", srv.FullyQualifiedName()))

	srvData := &ServiceData{
		Name:    m.ctx.Name(srv).String(),
		Methods: make([]*MethodData, 0, len(srv.Methods())),
	}

	// check service option: ServiceSkip
	srvSkip := false
	m.must(srv.Extension(redact.E_ServiceSkip, &srvSkip))
	if srvSkip {
		srvData.Skip = true
		m.Debug(fmt.Sprintf("Service %s is marked as skipped", srv.FullyQualifiedName()))
		// continue
	}

	// check internal service options
	srvInternal := false
	m.must(srv.Extension(redact.E_InternalService, &srvInternal))
	srvCode := uint32(codes.PermissionDenied) // default code
	if !m.must(srv.Extension(redact.E_InternalServiceCode, &srvCode)) {
		srvCode = uint32(codes.PermissionDenied)
	}

	// Validate status code with better error message
	if err := m.validateStatusCode(srvCode, srv.FullyQualifiedName()); err != nil {
		m.Fail(err)
		return nil
	}
	srvErrMsg := ""
	if !m.must(srv.Extension(redact.E_InternalServiceErrMessage, &srvErrMsg)) {
		srvErrMsg = defaultErrMsg
	}

	// methods
	for _, meth := range srv.Methods() {
		// Validate method before processing
		if err := m.validateMethod(meth); err != nil {
			m.Failf("Cannot process method %s: %v", meth.FullyQualifiedName(), err)
			continue
		}

		in := meth.Input()
		out := meth.Output()

		// Additional safety checks
		if in == nil {
			m.Failf("Method %s has nil input message", meth.FullyQualifiedName())
			continue
		}
		if out == nil {
			m.Failf("Method %s has nil output message", meth.FullyQualifiedName())
			continue
		}

		methData := &MethodData{
			Name:            m.ctx.Name(meth).String(),
			Input:           nameWithAlias(in),
			Output:          m.processMessage(out, nameWithAlias),
			ClientStreaming: meth.ClientStreaming(),
			ServerStreaming: meth.ServerStreaming(),
		}
		srvData.Methods = append(srvData.Methods, methData)

		// check method skip options
		methSkip := false
		m.must(meth.Extension(redact.E_MethodSkip, &methSkip))
		if methSkip || srvSkip {
			methData.Skip = true
			if methSkip {
				m.Debug(fmt.Sprintf("Method %s is marked as skipped", meth.FullyQualifiedName()))
			}
			continue
		}

		methInternal := false
		m.must(meth.Extension(redact.E_InternalMethod, &methInternal))
		methCode := srvCode // serviceCode
		if !m.must(meth.Extension(redact.E_InternalMethodCode, &methCode)) {
			methCode = srvCode
		}

		// Validate method status code with better error message
		if err := m.validateStatusCode(methCode, meth.FullyQualifiedName()); err != nil {
			m.Fail(err)
			continue
		}
		methErrMsg := srvErrMsg
		if !m.must(meth.Extension(redact.E_InternalMethodErrMessage, &methErrMsg)) {
			methErrMsg = srvErrMsg
		}

		// apply format specifiers
		methErrMsg = strings.ReplaceAll(methErrMsg, specifierMethod, methData.Name)
		methErrMsg = strings.ReplaceAll(methErrMsg, specifierService, srvData.Name)

		methData.ErrMessage = "`" + methErrMsg + "`"
		methData.StatusCode = codes.Code(methCode).String()
		methData.Internal = srvInternal || methInternal
	}
	return srvData
}

// processMessage extracts all pgs.Message and their pgs.Field(s) information and
// structures them into MessageData
func (m *Module) processMessage(
	msg pgs.Message,
	nameWithAlias func(n pgs.Entity) string,
	wantFields ...bool,
) *MessageData {
	// Validate message before processing
	if err := m.validateMessage(msg); err != nil {
		m.Failf("Cannot process message: %v", err)
		return nil
	}

	defer m.recoverFromPanic(fmt.Sprintf("processing message %s", msg.FullyQualifiedName()))

	msgData := &MessageData{
		Name:      m.ctx.Name(msg).String(),
		WithAlias: nameWithAlias(msg),
		Fields:    make([]*FieldData, 0, len(msg.Fields())*2),
	}

	// check message ignore options
	msgData.Ignore = false
	m.must(msg.Extension(redact.E_Ignored, &msgData.Ignore))
	if msgData.Ignore {
		m.Debug(fmt.Sprintf("Message %s is marked as ignored", msg.FullyQualifiedName()))
		return msgData
	}

	// check message nil options
	msgData.ToNil = false
	m.must(msg.Extension(redact.E_Nil, &msgData.ToNil))

	// check message empty options
	msgData.ToEmpty = false
	m.must(msg.Extension(redact.E_Empty, &msgData.ToEmpty))

	// Log warning if both nil and empty are set (validation should have caught this)
	if msgData.ToNil && msgData.ToEmpty {
		m.Debug(fmt.Sprintf("Warning: Message %s has both nil and empty options - this is invalid", msg.FullyQualifiedName()))
	}

	if len(wantFields) > 0 {
		for _, field := range msg.Fields() {
			msgData.Fields = append(msgData.Fields, m.processFields(field, nameWithAlias))
		}
	}
	return msgData
}
