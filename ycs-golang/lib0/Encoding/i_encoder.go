// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package encoding

// IEncoder is the interface that wraps the basic Write and encoding methods.
type IEncoder[T any] interface {
	Write(value T) error
	ToArray() ([]byte, error)
	GetBuffer() ([]byte, int, error)
	Dispose()
}
