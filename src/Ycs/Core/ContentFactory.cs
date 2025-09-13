// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using Ycs.Contracts;
using Ycs.Content;
using Ycs.Types;

namespace Ycs.Core
{
    /// <summary>
    /// Default implementation of content factory
    /// </summary>
    public class ContentFactory : IContentFactory
    {
        public IContent CreateContentType(IAbstractType type)
        {
            return new ContentType(type);
        }
        
        public IContent CreateContentDoc(object doc)
        {
            if (doc is YDoc yDoc)
            {
                return new ContentDoc(yDoc);
            }
            throw new ArgumentException("Expected YDoc instance", nameof(doc));
        }

        public IContent CreateContentBinary(byte[] ba)
        {
            return new ContentBinary(ba);
        }

        public IContent CreateContentAny(object value)
        {
            return new ContentAny(new object[] { value });
        }

        public IContent CreateContentFormat(string key, object value)
        {
            return new ContentFormat(key, value);
        }

        public IContent CreateContentString(string text)
        {
            return new ContentString(text);
        }

        public IContent CreateContentEmbed(object embed)
        {
            return new ContentEmbed(embed);
        }


        public IContent CreateContent(object value)
        {
            if (value == null)
            {
                return CreateContentAny(value);
            }
            switch (value)
            {
                case YDoc d:
                    return CreateContentDoc(d);
                case AbstractType at:
                    return CreateContentType(at);
                case byte[] ba:
                    return CreateContentBinary(ba);
                default:
                    return CreateContentAny(value);
            }
        }
        
 
    }
}