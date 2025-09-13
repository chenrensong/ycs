// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

namespace Ycs.Contracts
{
    /// <summary>
    /// Registry for type readers to avoid circular dependencies in ContentType.Read method
    /// </summary>
    public interface ITypeReaderRegistry
    {
        /// <summary>
        /// Register a type reader for a specific type reference ID
        /// </summary>
        void RegisterTypeReader(uint typeRefId, System.Func<IUpdateDecoder, IAbstractType> reader);
        
        /// <summary>
        /// Read AbstractType by type reference ID
        /// </summary>
        IAbstractType ReadType(uint typeRefId, IUpdateDecoder decoder);
    }
}