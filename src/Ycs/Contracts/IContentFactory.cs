// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

namespace Ycs.Contracts
{
    /// <summary>
    /// Factory interface for creating content objects without circular dependencies
    /// </summary>
    public interface IContentFactory
    {
        /// <summary>
        /// Creates ContentType for AbstractType instances
        /// </summary>
        IContent CreateContentType(IAbstractType type);
        
        /// <summary>
        /// Creates ContentDoc for YDoc instances
        /// </summary>
        IContent CreateContentDoc(object doc);

        IContent CreateContentBinary(byte[] ba);

        IContent CreateContentAny(object value);

        /// <summary>
        /// Creates appropriate content based on value type
        /// </summary>
        IContent CreateContent(object value);
        
        /// <summary>
        /// Creates ContentFormat for formatting
        /// </summary>
        IContent CreateContentFormat(string key, object value);
        
        /// <summary>
        /// Creates ContentString for string content
        /// </summary>
        IContent CreateContentString(string text);
        
        /// <summary>
        /// Creates ContentEmbed for embedded content
        /// </summary>
        IContent CreateContentEmbed(object embed);
    }
    
    /// <summary>
    /// Registry for content readers to avoid circular dependencies
    /// </summary>
    public interface IContentReaderRegistry
    {
        /// <summary>
        /// Register a content reader for a specific content reference ID
        /// </summary>
        void RegisterContentReader(int contentRef, System.Func<IUpdateDecoder, IContent> reader);
        
        /// <summary>
        /// Read content by content reference ID
        /// </summary>
        IContent ReadContent(int contentRef, IUpdateDecoder decoder);
    }
}