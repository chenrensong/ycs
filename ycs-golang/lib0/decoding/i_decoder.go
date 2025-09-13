// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package decoding

// IDecoder is an interface that represents a stream decoder
type IDecoder[T any] interface {
	Read() (T, error)
	Dispose()
}
