// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;
using Ycs.Contracts;

namespace Ycs.Core
{
    /// <summary>
    /// Default implementation of type reader registry
    /// </summary>
    public class TypeReaderRegistry : ITypeReaderRegistry
    {
        private readonly Dictionary<uint, Func<IUpdateDecoder, IAbstractType>> _readers = new Dictionary<uint, Func<IUpdateDecoder, IAbstractType>>();

        public void RegisterTypeReader(uint typeRefId, Func<IUpdateDecoder, IAbstractType> reader)
        {
            _readers[typeRefId] = reader;
        }

        public IAbstractType ReadType(uint typeRefId, IUpdateDecoder decoder)
        {
            if (_readers.TryGetValue(typeRefId, out var reader))
            {
                return reader(decoder);
            }
            
            throw new NotImplementedException($"Type {typeRefId} not implemented");
        }
    }
}