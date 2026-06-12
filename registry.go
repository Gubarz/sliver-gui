package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
)

type RegistryValue struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ListRegistrySubKeys returns the subkeys for a given hive and path
func (a *App) ListRegistrySubKeys(sessionID string, hive string, path string) (*sliverpb.RegistrySubKeyList, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}

	req := &sliverpb.RegistrySubKeyListReq{
		Request: &commonpb.Request{SessionID: sessionID},
		Hive:    hive,
		Path:    path,
	}

	return a.rpcClient.RegistryListSubKeys(context.Background(), req)
}

// ListRegistryValues returns the values for a given hive and path
func (a *App) ListRegistryValues(sessionID string, hive string, path string) (*sliverpb.RegistryValuesList, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}

	req := &sliverpb.RegistryListValuesReq{
		Request: &commonpb.Request{SessionID: sessionID},
		Hive:    hive,
		Path:    path,
	}

	return a.rpcClient.RegistryListValues(context.Background(), req)
}

func (a *App) ReadRegistryValue(sessionID, hive, path, key string) (*RegistryValue, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	resp, err := a.rpcClient.RegistryRead(context.Background(), &sliverpb.RegistryReadReq{
		Request: &commonpb.Request{SessionID: sessionID},
		Hive:    strings.ToUpper(hive),
		Path:    path,
		Key:     key,
	})
	if err != nil {
		return nil, err
	}
	if resp.Response != nil && resp.Response.Err != "" {
		return nil, fmt.Errorf("%s", resp.Response.Err)
	}
	return &RegistryValue{Name: key, Type: "Value", Value: resp.Value}, nil
}

func (a *App) WriteRegistryValue(sessionID, hive, path, key, valueType, value string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	req := &sliverpb.RegistryWriteReq{
		Request: &commonpb.Request{SessionID: sessionID},
		Hive:    strings.ToUpper(hive),
		Path:    path,
		Key:     key,
	}
	switch strings.ToLower(strings.TrimSpace(valueType)) {
	case "string":
		req.Type = sliverpb.RegistryTypeString
		req.StringValue = value
	case "binary":
		decoded, err := hex.DecodeString(strings.ReplaceAll(value, " ", ""))
		if err != nil {
			return fmt.Errorf("binary values must be hexadecimal: %w", err)
		}
		req.Type = sliverpb.RegistryTypeBinary
		req.ByteValue = decoded
	case "dword":
		parsed, err := strconv.ParseUint(value, 0, 32)
		if err != nil {
			return fmt.Errorf("invalid DWORD: %w", err)
		}
		req.Type = sliverpb.RegistryTypeDWORD
		req.DWordValue = uint32(parsed)
	case "qword":
		parsed, err := strconv.ParseUint(value, 0, 64)
		if err != nil {
			return fmt.Errorf("invalid QWORD: %w", err)
		}
		req.Type = sliverpb.RegistryTypeQWORD
		req.QWordValue = parsed
	default:
		return fmt.Errorf("unsupported registry value type %q", valueType)
	}
	resp, err := a.rpcClient.RegistryWrite(context.Background(), req)
	if err != nil {
		return err
	}
	if resp.Response != nil && resp.Response.Err != "" {
		return fmt.Errorf("%s", resp.Response.Err)
	}
	return nil
}

func (a *App) CreateRegistryKey(sessionID, hive, path, key string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	resp, err := a.rpcClient.RegistryCreateKey(context.Background(), &sliverpb.RegistryCreateKeyReq{
		Request: &commonpb.Request{SessionID: sessionID},
		Hive:    strings.ToUpper(hive),
		Path:    path,
		Key:     key,
	})
	if err != nil {
		return err
	}
	if resp.Response != nil && resp.Response.Err != "" {
		return fmt.Errorf("%s", resp.Response.Err)
	}
	return nil
}

// DeleteRegistryEntry removes a named value, or a subkey when no value with
// that name exists. This matches Sliver's RegistryDeleteKey RPC semantics.
func (a *App) DeleteRegistryEntry(sessionID, hive, path, key string) error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}
	resp, err := a.rpcClient.RegistryDeleteKey(context.Background(), &sliverpb.RegistryDeleteKeyReq{
		Request: &commonpb.Request{SessionID: sessionID},
		Hive:    strings.ToUpper(hive),
		Path:    path,
		Key:     key,
	})
	if err != nil {
		return err
	}
	if resp.Response != nil && resp.Response.Err != "" {
		return fmt.Errorf("%s", resp.Response.Err)
	}
	return nil
}
