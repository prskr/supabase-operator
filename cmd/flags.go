/*
Copyright 2025 Peter Kurfer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var _ kong.MapperValue = (*FileContent)(nil)

type FileContent []byte

func (f *FileContent) Decode(ctx *kong.DecodeContext) (err error) {
	var filePath string
	if err := ctx.Scan.PopValueInto("file-content", &filePath); err != nil {
		return err
	}

	if *f, err = os.ReadFile(filePath); err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return nil
}
