// +build !amd64 go1.21

/*
* Copyright 2023 ByteDance Inc.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package decoder

import (
     `encoding/json`
     `bytes`
     `reflect`
     `github.com/bytedance/sonic/internal/native/types`
     `github.com/bytedance/sonic/option`
     `io`
)

const (
     _F_use_int64 = iota
     _F_use_number
     _F_disable_urc
     _F_disable_unknown
     _F_copy_string
     _F_validate_string

     _F_allow_control = 31
)

type Options uint64

const (
     OptionUseInt64         Options = 1 << _F_use_int64
     OptionUseNumber        Options = 1 << _F_use_number
     OptionUseUnicodeErrors Options = 1 << _F_disable_urc
     OptionDisableUnknown   Options = 1 << _F_disable_unknown
     OptionCopyString       Options = 1 << _F_copy_string
     OptionValidateString   Options = 1 << _F_validate_string
)

func (self *Decoder) SetOptions(opts Options) {
     if (opts & OptionUseNumber != 0) && (opts & OptionUseInt64 != 0) {
         panic("can't set OptionUseInt64 and OptionUseNumber both!")
     }
     self.f = uint64(opts)
}


// Decoder is the decoder context object
type Decoder struct {
     i int
     f uint64
     s string
}

// NewDecoder creates a new decoder instance.
func NewDecoder(s string) *Decoder {
     return &Decoder{s: s}
}

// Pos returns the current decoding position.
func (self *Decoder) Pos() int {
     return self.i
}

func (self *Decoder) Reset(s string) {
     self.s = s
     self.i = 0
     // self.f = 0
}

// NOTE: api fallback do nothing
func (self *Decoder) CheckTrailings() error {
     pos := self.i
     buf := self.s
     /* skip all the trailing spaces */
     if pos != len(buf) {
         for pos < len(buf) && (types.SPACE_MASK & (1 << buf[pos])) != 0 {
             pos++
         }
     }

     /* then it must be at EOF */
     if pos == len(buf) {
         return nil
     }

     /* junk after JSON value */
     return nil
}


// Decode parses the JSON-encoded data from current position and stores the result
// in the value pointed to by val.
func (self *Decoder) Decode(val interface{}) error {
    r := bytes.NewBufferString(self.s)
   dec := json.NewDecoder(r)
   if (self.f | uint64(OptionUseNumber)) != 0  {
       dec.UseNumber()
   }
   if (self.f | uint64(OptionDisableUnknown)) != 0  {
       dec.DisallowUnknownFields()
   }
   return dec.Decode(val)
}

// UseInt64 indicates the Decoder to unmarshal an integer into an interface{} as an
// int64 instead of as a float64.
func (self *Decoder) UseInt64() {
     self.f  |= 1 << _F_use_int64
     self.f &^= 1 << _F_use_number
}

// UseNumber indicates the Decoder to unmarshal a number into an interface{} as a
// json.Number instead of as a float64.
func (self *Decoder) UseNumber() {
     self.f &^= 1 << _F_use_int64
     self.f  |= 1 << _F_use_number
}

// UseUnicodeErrors indicates the Decoder to return an error when encounter invalid
// UTF-8 escape sequences.
func (self *Decoder) UseUnicodeErrors() {
     self.f |= 1 << _F_disable_urc
}

// DisallowUnknownFields indicates the Decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported fields in the destination.
func (self *Decoder) DisallowUnknownFields() {
     self.f |= 1 << _F_disable_unknown
}

// CopyString indicates the Decoder to decode string values by copying instead of referring.
func (self *Decoder) CopyString() {
     self.f |= 1 << _F_copy_string
}

// ValidateString causes the Decoder to validate string values when decoding string value 
// in JSON. Validation is that, returning error when unescaped control chars(0x00-0x1f) or
// invalid UTF-8 chars in the string value of JSON.
func (self *Decoder) ValidateString() {
     self.f |= 1 << _F_validate_string
}

// Pretouch compiles vt ahead-of-time to avoid JIT compilation on-the-fly, in
// order to reduce the first-hit latency.
//
// Opts are the compile options, for example, "option.WithCompileRecursiveDepth" is
// a compile option to set the depth of recursive compile for the nested struct type.
func Pretouch(vt reflect.Type, opts ...option.CompileOption) error {
     return nil
}

type StreamDecoder struct {
   r       io.Reader
   buf     []byte
   scanp   int
   scanned int64
   err     error
   Decoder
}

// NewStreamDecoder adapts to encoding/json.NewDecoder API.
//
// NewStreamDecoder returns a new decoder that reads from r.
func NewStreamDecoder(r io.Reader) *StreamDecoder {
   return &StreamDecoder{r : r}
}

// Decode decodes input stream into val with corresponding data. 
// Redundantly bytes may be read and left in its buffer, and can be used at next call.
// Either io error from underlying io.Reader (except io.EOF) 
// or syntax error from data will be recorded and stop subsequently decoding.
func (self *StreamDecoder) Decode(val interface{}) (err error) {
   dec := json.NewDecoder(self.r)
   if (self.f | uint64(OptionUseNumber)) != 0  {
       dec.UseNumber()
   }
   if (self.f | uint64(OptionDisableUnknown)) != 0  {
       dec.DisallowUnknownFields()
   }
   return dec.Decode(val)
}

