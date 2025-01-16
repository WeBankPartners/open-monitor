/*
 * Copyright 2021 ByteDance Inc.
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

package encoder

import (
    `encoding/json`
    `io`
)

// StreamEncoder uses io.Writer as input.
type StreamEncoder struct {
    w io.Writer
    Encoder
}

// NewStreamEncoder adapts to encoding/json.NewDecoder API.
//
// NewStreamEncoder returns a new encoder that write to w.
func NewStreamEncoder(w io.Writer) *StreamEncoder {
    return &StreamEncoder{w: w}
}

// Encode encodes interface{} as JSON to io.Writer
func (enc *StreamEncoder) Encode(val interface{}) (err error) {
    out := newBytes()

    /* encode into the buffer */
    err = EncodeInto(&out, val, enc.Opts)
    if err != nil {
        goto free_bytes
    }

    if enc.indent != "" || enc.prefix != "" {
        /* indent the JSON */
        buf := newBuffer()
        err = json.Indent(buf, out, enc.prefix, enc.indent)
        if err != nil {
            freeBuffer(buf)
            goto free_bytes
        }

        // according to standard library, terminate each value with a newline...
        buf.WriteByte('\n')

        /* copy into io.Writer */
        _, err = io.Copy(enc.w, buf)
        if err != nil {
            freeBuffer(buf)
            goto free_bytes
        }

    } else {
        /* copy into io.Writer */
        var n int
        for len(out) > 0 {
            n, err = enc.w.Write(out)
            out = out[n:]
            if err != nil {
                goto free_bytes
            }
        }

        // according to standard library, terminate each value with a newline...
        enc.w.Write([]byte{'\n'})
    }

free_bytes:
    freeBytes(out)
    return err
}