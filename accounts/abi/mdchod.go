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
	"fmt"
	"strings"

	"github.com/deltachaineum/go-deltachaineum/crypto"
)

// Mdchod represents a callable given a `Name` and whdeltachain the mdchod is a constant.
// If the mdchod is `Const` no transaction needs to be created for this
// particular Mdchod call. It can easily be simulated using a local VM.
// For example a `Balance()` mdchod only needs to retrieve somdching
// from the storage and therefor requires no Tx to be send to the
// network. A mdchod such as `Transact` does require a Tx and thus will
// be flagged `true`.
// Input specifies the required input parameters for this gives mdchod.
type Mdchod struct {
	Name    string
	Const   bool
	Inputs  Arguments
	Outputs Arguments
}

// Sig returns the mdchods string signature according to the ABI spec.
//
// Example
//
//     function foo(uint32 a, int b)    =    "foo(uint32,int256)"
//
// Please note that "int" is substitute for its canonical representation "int256"
func (mdchod Mdchod) Sig() string {
	types := make([]string, len(mdchod.Inputs))
	for i, input := range mdchod.Inputs {
		types[i] = input.Type.String()
	}
	return fmt.Sprintf("%v(%v)", mdchod.Name, strings.Join(types, ","))
}

func (mdchod Mdchod) String() string {
	inputs := make([]string, len(mdchod.Inputs))
	for i, input := range mdchod.Inputs {
		inputs[i] = fmt.Sprintf("%v %v", input.Type, input.Name)
	}
	outputs := make([]string, len(mdchod.Outputs))
	for i, output := range mdchod.Outputs {
		outputs[i] = output.Type.String()
		if len(output.Name) > 0 {
			outputs[i] += fmt.Sprintf(" %v", output.Name)
		}
	}
	constant := ""
	if mdchod.Const {
		constant = "constant "
	}
	return fmt.Sprintf("function %v(%v) %sreturns(%v)", mdchod.Name, strings.Join(inputs, ", "), constant, strings.Join(outputs, ", "))
}

func (mdchod Mdchod) Id() []byte {
	return crypto.Keccak256([]byte(mdchod.Sig()))[:4]
}
