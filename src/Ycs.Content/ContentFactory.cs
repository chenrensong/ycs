// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using Ycs.Contracts;

namespace Ycs.Content
{
    /// <summary>
    /// Default implementation of content factory
    /// </summary>
    public class ContentFactory
    {
        public static IContent CreateContentType(IAbstractType type)
        {
            return new ContentType(type);
        }
        
        public static IContent CreateContentDoc(object doc)
        {
            if (doc is IYDoc yDoc)
            {
                return new ContentDoc(yDoc);
            }
            throw new ArgumentException("Expected YDoc instance", nameof(doc));
        }

        public static IContent CreateContentBinary(byte[] ba)
        {
            return new ContentBinary(ba);
        }

        public static IContent CreateContentAny(object value)
        {
            return new ContentAny(new object[] { value });
        }

        public static IContent CreateContentFormat(string key, object value)
        {
            return new ContentFormat(key, value);
        }

        public static IContent CreateContentString(string text)
        {
            return new ContentString(text);
        }

        public static IContent CreateContentEmbed(object embed)
        {
            return new ContentEmbed(embed);
        }

        public static IContent CreateContent(object value)
        {
            if (value == null)
            {
                return CreateContentAny(value);
            }
            switch (value)
            {
                case IYDoc d:
                    return CreateContentDoc(d);
                case IAbstractType at:
                    return CreateContentType(at);
                case byte[] ba:
                    return CreateContentBinary(ba);
                default:
                    return CreateContentAny(value);
            }
        }
        
 
    }
}