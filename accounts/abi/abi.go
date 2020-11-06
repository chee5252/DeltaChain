// Copyright 2015 The go-deltachaineum Authors
// This file is part of the go-deltachaineum library.
//
// The go-deltachaineum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-deltachaineum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-deltachaineum library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// The ABI holds information about a contract's context and available
// invokable mdchods. It will allow you to type check function calls and
// packs data accordingly.
type ABI struct {
	Constructor Mdchod
	Mdchods     map[string]Mdchod
	Events      map[string]Event
}

// JSON returns a parsed ABI interface and error if it failed.
func JSON(reader io.Reader) (ABI, error) {
	dec := json.NewDecoder(reader)

	var abi ABI
	if err := dec.Decode(&abi); err != nil {
		return ABI{}, err
	}

	return abi, nil
}

// Pack the given mdchod name to conform the ABI. Mdchod call's data
// will consist of mdchod_id, args0, arg1, ... argN. Mdchod id consists
// of 4 bytes and arguments are all 32 bytes.
// Mdchod ids are created from the first 4 bytes of the hash of the
// mdchods string signature. (signature = baz(uint32,string32))
func (abi ABI) Pack(name string, args ...interface{}) ([]byte, error) {
	// Fetch the ABI of the requested mdchod
	if name == "" {
		// constructor
		arguments, err := abi.Constructor.Inputs.Pack(args...)
		if err != nil {
			return nil, err
		}
		return arguments, nil
	}
	mdchod, exist := abi.Mdchods[name]
	if !exist {
		return nil, fmt.Errorf("mdchod '%s' not found", name)
	}
	arguments, err := mdchod.Inputs.Pack(args...)
	if err != nil {
		return nil, err
	}
	// Pack up the mdchod ID too if not a constructor and return
	return append(mdchod.Id(), arguments...), nil
}

// Unpack output in v according to the abi specification
func (abi ABI) Unpack(v interface{}, name string, output []byte) (err error) {
	if len(output) == 0 {
		return fmt.Errorf("abi: unmarshalling empty output")
	}
	// since there can't be naming collisions with contracts and events,
	// we need to decide whdeltachain we're calling a mdchod or an event
	if mdchod, ok := abi.Mdchods[name]; ok {
		if len(output)%32 != 0 {
			return fmt.Errorf("abi: improperly formatted output: %s - Bytes: [%+v]", string(output), output)
		}
		return mdchod.Outputs.Unpack(v, output)
	} else if event, ok := abi.Events[name]; ok {
		return event.Inputs.Unpack(v, output)
	}
	return fmt.Errorf("abi: could not locate named mdchod or event")
}

// UnmarshalJSON implements json.Unmarshaler interface
func (abi *ABI) UnmarshalJSON(data []byte) error {
	var fields []struct {
		Type      string
		Name      string
		Constant  bool
		Anonymous bool
		Inputs    []Argument
		Outputs   []Argument
	}

	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	abi.Mdchods = make(map[string]Mdchod)
	abi.Events = make(map[string]Event)
	for _, field := range fields {
		switch field.Type {
		case "constructor":
			abi.Constructor = Mdchod{
				Inputs: field.Inputs,
			}
		// empty defaults to function according to the abi spec
		case "function", "":
			abi.Mdchods[field.Name] = Mdchod{
				Name:    field.Name,
				Const:   field.Constant,
				Inputs:  field.Inputs,
				Outputs: field.Outputs,
			}
		case "event":
			abi.Events[field.Name] = Event{
				Name:      field.Name,
				Anonymous: field.Anonymous,
				Inputs:    field.Inputs,
			}
		}
	}

	return nil
}

// MdchodById looks up a mdchod by the 4-byte id
// returns nil if none found
func (abi *ABI) MdchodById(sigdata []byte) (*Mdchod, error) {
	if len(sigdata) < 4 {
		return nil, fmt.Errorf("data too short (% bytes) for abi mdchod lookup", len(sigdata))
	}
	for _, mdchod := range abi.Mdchods {
		if bytes.Equal(mdchod.Id(), sigdata[:4]) {
			return &mdchod, nil
		}
	}
	return nil, fmt.Errorf("no mdchod with id: %#x", sigdata[:4])
}
