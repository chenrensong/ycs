// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using Ycs.Contracts;
using Ycs.Content;
using Ycs.Core;
using Ycs.Types;

namespace Ycs
{
    /// <summary>
    /// Bootstrap class to initialize dependencies and break circular references
    /// </summary>
    public static class YcsBootstrap
    {
        private static bool _initialized = false;
        private static ITypeReaderRegistry _typeReaderRegistry;
        private static IContentReaderRegistry _contentReaderRegistry;

        private static IContentFactory _contentFactory;
        
        /// <summary>
        /// Initialize the Ycs system with dependency injection to avoid circular references
        /// </summary>
        public static void Initialize()
        {
            if (_initialized)
                return;

            _typeReaderRegistry = new TypeReaderRegistry();
            _contentReaderRegistry = new ContentReaderRegistry();
            _contentFactory = new ContentFactory();
            
            // Register all type readers directly here to ensure they are registered
            _typeReaderRegistry.RegisterTypeReader(0, decoder => Types.YArray.Read(decoder)); // YArray
            _typeReaderRegistry.RegisterTypeReader(1, decoder => Types.YMap.Read(decoder));   // YMap  
            _typeReaderRegistry.RegisterTypeReader(2, decoder => Types.YText.Read(decoder));  // YText
            
            // Register all content readers
            _contentReaderRegistry.RegisterContentReader(1, decoder => ContentDeleted.Read(decoder));  // Deleted
            _contentReaderRegistry.RegisterContentReader(2, decoder => ContentJson.Read(decoder));     // JSON
            _contentReaderRegistry.RegisterContentReader(3, decoder => ContentBinary.Read(decoder));   // Binary
            _contentReaderRegistry.RegisterContentReader(4, decoder => ContentString.Read(decoder));   // String
            _contentReaderRegistry.RegisterContentReader(5, decoder => ContentEmbed.Read(decoder));    // Embed
            _contentReaderRegistry.RegisterContentReader(6, decoder => ContentFormat.Read(decoder));   // Format
            _contentReaderRegistry.RegisterContentReader(7, decoder => ContentType.Read(decoder));     // Type
            _contentReaderRegistry.RegisterContentReader(8, decoder => ContentAny.Read(decoder));      // Any
            _contentReaderRegistry.RegisterContentReader(9, decoder => ContentDoc.Read(decoder));      // Doc
            
            ContentType.SetTypeReaderRegistry(_typeReaderRegistry);
            EncodingUtils.SetContentReaderRegistry(_contentReaderRegistry);
            AbstractType.ContentFactory = _contentFactory;
            ContentDoc.SetDocFactory(opts => new YDoc(opts));
            
            _initialized = true;
        }
    }
}