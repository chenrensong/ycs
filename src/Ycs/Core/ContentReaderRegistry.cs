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
    /// Default implementation of content reader registry
    /// </summary>
    public class ContentReaderRegistry : IContentReaderRegistry
    {
        private readonly Dictionary<int, Func<IUpdateDecoder, IContent>> _readers = new Dictionary<int, Func<IUpdateDecoder, IContent>>();

        public void RegisterContentReader(int contentRef, Func<IUpdateDecoder, IContent> reader)
        {
            _readers[contentRef] = reader;
        }

        public IContent ReadContent(int contentRef, IUpdateDecoder decoder)
        {
            if (_readers.TryGetValue(contentRef, out var reader))
            {
                return reader(decoder);
            }
            
            throw new InvalidOperationException($"Content type not recognized: {contentRef}");
        }
    }
}