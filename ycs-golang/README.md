# YCS Golang Contracts

This directory contains the Go translation of the C# YCS (Y-CRDT) contracts from the original C# implementation.

## Structure

The contracts are organized into the following files:

### Core Types
- `struct_id.go` - StructID type and related functionality
- `delete_item.go` - DeleteItem struct
- `changes_collection.go` - ChangesCollection and related types (Delta, ChangeAction, ChangeKey)
- `ydoc_options.go` - YDocOptions configuration struct

### Interfaces
- `interfaces.go` - Core interfaces (IStructItem, IAbstractType, IContent, IContentEx)
- `update_encoder.go` - Encoder interfaces (IDSEncoder, IUpdateEncoder)
- `update_decoder.go` - Decoder interfaces (IDSDecoder, IUpdateDecoder)
- `transaction.go` - ITransaction interface
- `struct_store.go` - IStructStore interface
- `delete_set.go` - IDeleteSet interface
- `snapshot.go` - ISnapshot interface

### Y-Type Interfaces
- `yarray.go` - IYArray and IYArrayBase interfaces
- `ymap.go` - IYMap interface
- `ytext.go` - IYText interface and related types (YTextChangeType, YTextChangeAttributes)
- `yevent.go` - IYEvent interface
- `ydoc.go` - IYDoc interface and event handler types

### Factory and Registry Interfaces
- `type_reader_registry.go` - ITypeReaderRegistry interface
- `content_factory.go` - IContentFactory and IContentReaderRegistry interfaces

## Translation Notes

### Key Differences from C#

1. **Interface Naming**: Go interfaces follow Go conventions (e.g., `IStructItem` instead of `IStructItem`)

2. **Properties**: C# properties are translated to getter/setter method pairs in Go interfaces

3. **Optional Parameters**: C# optional parameters are implemented using Go variadic parameters with default handling in implementations

4. **Events**: C# events are translated to handler registration methods in Go

5. **Generics**: C# generics are translated to use `interface{}` or function parameters where appropriate

6. **Collections**: 
   - `ISet<T>` → `map[T]struct{}`
   - `IList<T>` → `[]T`
   - `IDictionary<K,V>` → `map[K]V`

7. **Nullable Types**: C# nullable types (`T?`) are translated to pointers (`*T`) in Go

8. **Error Handling**: Go uses explicit error returns instead of exceptions

### Implementation Requirements

Implementations of these interfaces should:

1. Handle default parameter values for variadic parameters
2. Implement proper error handling
3. Follow Go naming conventions
4. Use appropriate Go data structures
5. Handle nil pointer cases safely

## Usage

This package provides the contract definitions. Concrete implementations should be provided in separate packages that import these contracts.

```go
import "ycs-golang/contracts"

// Example implementation
type MyYArray struct {
    // implementation fields
}

func (ya *MyYArray) Add(content []interface{}) {
    // implementation
}

// ... other interface methods
``` 